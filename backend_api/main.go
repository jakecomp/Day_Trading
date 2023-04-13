package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
)

var db *mongo.Client
var ctx context.Context
var rabbitChannel *amqp.Channel
var queue amqp.Queue
var upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 0,
}

const database = "day_trading"

type Credentials struct {
	Password string `json:"password", db:"password"`
	Username string `json:"username", db:"username"`
}

type user_doc struct {
	Username string
	Hash     string
	Balance  float32
	Stonks   map[string]int
}

type command struct {
	Ticket  int
	Command string
	Args    []string
}

type Message struct {
	Command string
	Data    *command
}

type quote struct {
	Stock string  `json:"stock"`
	Price float64 `json:"price"`
}

type quote_log struct {
	Timestamp    int64  `xml:"timestamp"`
	Username     string `xml:"username" json:"username"`
	Ticketnumber int    `xml:"ticketnumber" json:"ticketnumber"`
	Price        string `xml:"price" json:"price"`
	StockSymbol  string `xml:"stock_symbol" json:"stock_symbol"`
}

type StockHolder struct {
	stocks map[string]quote
	lock   sync.RWMutex
}

func (sh *StockHolder) GetStock(stock string) (quote, bool) {
	sh.lock.RLock()
	p, ok := sh.stocks[stock]
	sh.lock.RUnlock()
	return p, ok
}
func (sh *StockHolder) SetStock(stock string, q quote) {
	sh.lock.Lock()
	sh.stocks[stock] = q
	sh.lock.Unlock()
}

var stocks_map = StockHolder{
	stocks: make(map[string]quote),
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("signup hit")
	setupCORS(w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)

	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// Salt and hash the password using the bcrypt algorithm
	hashedPassword := hashPassword(creds.Password)

	// Create User Doc for DB
	var doc = new(user_doc)
	doc.Username = creds.Username
	doc.Hash = string(hashedPassword)
	doc.Balance = 0
	doc.Stonks = make(map[string]int)

	// Save User Doc to MongoDB
	if db == nil {
		db, ctx = connect()
	}
	collection := db.Database(database).Collection("users")
	_, err = collection.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("Error Inserting to DB: ", err)
		db.Disconnect(ctx)
		return
	}

	fmt.Fprintf(w, "SUCCESS")
}

func Signin(w http.ResponseWriter, r *http.Request) {
	setupCORS(w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	log.Println("signin hit")
	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Grab Stored Hash and Compare
	var result user_doc
	if db == nil {
		db, ctx = connect()
	}

	err = db.Database(database).Collection("users").FindOne(ctx, bson.D{{"username", creds.Username}}).Decode(&result)
	// db.Disconnect(ctx)
	if err != nil {
		fmt.Println("Error search for record: ", err)
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Hash), []byte(creds.Password))
	if err != nil {
		fmt.Println("Login or Password is incorrect!")
		panic(err)
	}

	fmt.Println("Verified user " + creds.Username)

	// Generate JWT and Return It
	token, err := generateJWT(creds.Username)
	if err != nil {
		fmt.Println("Error generating JWT: ", err)
		panic(err)
	}

	fmt.Fprintf(w, string(token))
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate User
	var valid = validateToken(w, r)
	if valid != nil {
		fmt.Println("Token not valid! Error: ", valid)
		return
	}

	// Upgrade our raw HTTP connection to a websocket based one
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()
	socketReader(conn)
}

func socketReader(conn *websocket.Conn) {
	// Event Loop, Handle Comms in here
	log.Println("Waiting for messages...")
	var cmds []command
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error during message reading:", err)
			break
		}
		// Pass the messages along to the Worker to keep the queue saturated
		err = rabbitChannel.Publish(
			"",         // Exchange name
			queue.Name, // Queue name
			false,      // Mandatory
			false,      // Immediate
			amqp.Publishing{
				ContentType: "text/json",
				Body:        message,
			},
		)
		if err != nil {
			fmt.Println("Error during message writing:", err)
			break
		}

		err = json.Unmarshal(message, &cmds)
		if err != nil {
			fmt.Println("JSON: ", cmds[0])
		}

		// Identify the important commands now
		for _, cmd := range cmds {
			// Check if should queue item
			if cmd.Command == "DUMPLOG" {
				fmt.Printf("DUMPLOG FOUND")
				//_, err := http.Post("http://10.9.0.9:8004/userlog", "application/json", "")
				//if err != nil {
				//	log.Fatal(err)
				//}
			} else if cmd.Command == "QUOTE" {
				// Get a quote
				var new_quote quote
				// DO WE KNOW THIS STOCK?
				if _, ok := stocks_map.GetStock(cmd.Args[1]); !ok {
					// println("Stock is not in map, updating map...")
					requestStockPrice(cmd.Args[1])
				}

				new_quote.Stock = cmd.Args[1]
				tmp, _ := stocks_map.GetStock(new_quote.Stock)
				new_quote.Price = tmp.Price

				log := quote_log{
					Timestamp:    time.Now().Unix(),
					Username:     cmd.Args[0],
					Ticketnumber: cmd.Ticket,
					Price:        fmt.Sprintf("%v", new_quote.Price),
					StockSymbol:  new_quote.Stock,
				}

				log_bytes, err := json.Marshal(log)

				_, err = http.Post("http://10.9.0.9:8004/quotelog", "application/json", bytes.NewBuffer(log_bytes))
				if err != nil {
					// fmt.Println(err)
				}
			} else if stringInSlice(cmd.Command, []string{"SET_BUY_AMOUNT", "SET_SELL_AMOUNT", "SET_BUY_TRIGGER", "SET_SELL_TRIGGER", "CANCEL_SET_BUY", "CANCEL_SET_SELL"}) {

				if !stringInSlice(cmd.Command, []string{"CANCEL_BUY", "CANCEL_SELL"}) {
					// DO WE KNOW THIS STOCK?
					if _, ok := stocks_map.GetStock(cmd.Args[1]); !ok {
						// println("Stock is not in map, updating map...")
						requestStockPrice(cmd.Args[1])
					}
				}

				msg, _ := json.Marshal(cmd)
				err = rabbitChannel.Publish(
					"",        // Exchange name
					"trigger", // Queue name
					false,     // Mandatory
					false,     // Immediate
					amqp.Publishing{
						ContentType: "text/json",
						Body:        []byte(msg),
					},
				)
				// println(" [x] Sent Trigger %s", msg)
				failOnError(err, "Failed to publish a message")
			} else {
				// fmt.Printf("Received Unknown Command: %s\n", cmd.Command)
			}
		}
		// This Should return success or failure eventually
		// err = conn.WriteMessage(messageType, message)
		// if err != nil {
		// 	fmt.Println("Error during message writing:", err)
		// 	break
		// }
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func setupCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/signup", Signup)
	http.HandleFunc("/signin", Signin)
	http.HandleFunc("/ws", socketHandler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
func dial(url string) (*amqp.Connection, error) {
	for {
		conn, err := amqp.Dial("amqp://guest:guest@10.9.0.10:5672/")
		if err == nil {
			return conn, err
		}
		// Rabbitmq is slow to start so we might have to wait on it
		time.Sleep(time.Second * 3)
	}
}

func main() {
	db, ctx = connect()
	defer db.Disconnect(ctx)

	log.SetOutput(ioutil.Discard)

	// Connect to RabbitMQ server
	time.Sleep(time.Second * 15)
	conn, err := dial("amqp://guest:guest@10.9.0.10:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	queue, rabbitChannel = connectQueue(conn)
	defer rabbitChannel.Close()

	setupStockListener()

	log.Println("RUNNING ON PORT 8000...")
	handleRequests()
}

func requestStockPrice(stock string) {
	err := rabbitChannel.Publish(
		"",               // name
		"stock_requests", // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(stock),
		})
	if err != nil {
		log.Println(err)
	}
}

func setupStockListener() {

	// Queue for recieving stock prices
	err := rabbitChannel.ExchangeDeclare(
		"stock_prices", // name
		"fanout",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Println("Failed to declare an exchange")
		panic(err)
	}
	// Queue for requesting stocks from queuer
	_, err = rabbitChannel.QueueDeclare(
		"stock_requests", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)

	q, err := rabbitChannel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Println("Failed to declare a queue")
		panic(err)
	}

	// Bind our temperary queue to the global exchange (subscribe to stock prices)
	err = rabbitChannel.QueueBind(
		q.Name,         // queue name
		"",             // routing key
		"stock_prices", // exchange
		false,
		nil,
	)
	if err != nil {
		log.Println("Failed to bind a queue")
		panic(err)
	}
	failOnError(err, "Failed to bind a queue")

	msgs, err := rabbitChannel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// This routine updates local map of stocks with the queuer that publishes every second
	go func() {
		for d := range msgs {
			json.Unmarshal(d.Body, &stocks_map)
		}
	}()
}

func connect() (*mongo.Client, context.Context) {
	clientOptions := options.Client()
	clientOptions.ApplyURI("mongodb://admin:admin@10.9.0.3:27017")
	// clientOptions.ApplyURI("mongodb://admin:admin@localhost:27017")
	ctx := context.Background()
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		fmt.Println("Error connecting to DB")
		panic(err)
	}
	return client, ctx
}

func connectQueue(conn *amqp.Connection) (amqp.Queue, *amqp.Channel) {

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		"trigger", // Queue name
		false,     // Durable
		false,     // Delete when unused
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Declare a queue
	q, err = ch.QueueDeclare(
		"worker", // Queue name
		false,    // Durable
		false,    // Delete when unused
		false,    // Exclusive
		false,    // No-wait
		nil,      // Arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q, ch
}

func hashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		fmt.Println("Error hashing: ", err)
		panic(err)
	}
	return string(hashedPassword)
}

var sampleSecretKey = []byte("THISISASECRETKEY")

func generateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(sampleSecretKey)

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}

func validateToken(w http.ResponseWriter, r *http.Request) (err error) {

	test := r.URL.Query().Get("token")
	fmt.Println("Received a token: ", test)

	token, err := jwt.Parse(r.URL.Query().Get("token"), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing")
		}
		return sampleSecretKey, nil
	})

	if token == nil {
		fmt.Fprintf(w, "invalid token")
		return errors.New("Token error, invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Fprintf(w, "couldn't parse claims")
		return errors.New("Token error, claims error")
	}

	exp := claims["exp"].(float64)
	if int64(exp) < time.Now().Local().Unix() {
		fmt.Fprintf(w, "token expired")
		return errors.New("Token error")
	}

	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
