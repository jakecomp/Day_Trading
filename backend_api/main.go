package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var db *mongo.Client
var ctx context.Context
var queueServiceConn *amqp.Channel
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

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: signup")
	setupCORS(w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	fmt.Println(creds)
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
	result, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("Error Inserting to DB: ", err)
		db.Disconnect(ctx)
		return
	}

	fmt.Fprintf(w, "SIGNED YOU UP!")
	fmt.Println("signup user " + doc.Username + " with hash " + doc.Hash)
	print(result)
}

func Signin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: signin")
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
	fmt.Println("Token generated: ", string(token))
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate User
	fmt.Println("Endpoint Hit: WS")
	var valid = validateToken(w, r)
	if valid != nil {
		fmt.Println("Token not valid! Error: ", valid)
		return
	}

	fmt.Println("Token Valid! Connecting Websocket...")

	// Upgrade our raw HTTP connection to a websocket based one
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()
	//err = conn.WriteMessage(1, []byte("Hello there"))
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
		fmt.Printf("Received: %s", string(message))

		err = json.Unmarshal(message, cmds)
		fmt.Println("JSON: ", cmds)
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
				allquote := make(map[string]quote)
				var thisStock quote
				resp, err := http.Get("http://10.9.0.6:8002/" + cmd.Args[1])
				if resp.StatusCode != http.StatusOK {
					fmt.Println("Error: Failed to get quote. ", err)

				}
				json.NewDecoder(resp.Body).Decode(&thisStock)
				allquote[thisStock.Stock] = thisStock
				log := quote_log{
					Timestamp:    time.Now().Unix(),
					Username:     cmd.Args[0],
					Ticketnumber: cmd.Ticket,
					Price:        fmt.Sprintf("%v", thisStock.Price),
					StockSymbol:  thisStock.Stock,
				}

				log_bytes, err := json.Marshal(log)

				go http.Post("http://10.9.0.9:8004/quotelog", "application/json", bytes.NewBuffer(log_bytes))
			}

		}

		msg, _ := json.Marshal(cmds)
		err = queueServiceConn.Publish(
			"",         // Exchange name
			queue.Name, // Queue name
			false,      // Mandatory
			false,      // Immediate
			amqp.Publishing{
				ContentType: "text/json",
				Body:        message,
			},
		)
		log.Printf(" [x] Sent %s", msg)
		failOnError(err, "Failed to publish a message")

		if err != nil {
			fmt.Println("Error during message writing:", err)
			break
		}

		// This Should return success or failure eventually
		// err = conn.WriteMessage(messageType, message)
		if err != nil {
			fmt.Println("Error during message writing:", err)
			break
		}
	}
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

	queue, queueServiceConn = connectQueue(conn)
	defer queueServiceConn.Close()
	fmt.Print("queue created")
	log.Println("RUNNING ON PORT 8000...")
	handleRequests()
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

	// Declare a queue
	q, err := ch.QueueDeclare(
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
