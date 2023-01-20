package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var db *mongo.Client
var upgrader = websocket.Upgrader{}

const database = "day_trading"

type Credentials struct {
	Password string `json:"password", db:"password"`
	Username string `json:"username", db:"username"`
}

type user_doc struct {
	Username string
	Hash     string
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: signup")

	db, ctx := connect()

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

	// Save User Doc to MongoDB
	collection := db.Database(database).Collection("users")
	result, err := collection.InsertOne(context.TODO(), doc)
	db.Disconnect(ctx)
	if err != nil {
		fmt.Println("Error Inserting to DB: ", err)
		return
	}

	fmt.Fprintf(w, "SIGNED YOU UP!")
	fmt.Println("signup user " + doc.Username + " with hash " + doc.Hash)
	print(result)
}

func Signin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: signin")
	db, ctx := connect()

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
	err = db.Database(database).Collection("users").FindOne(ctx, bson.D{{"username", creds.Username}}).Decode(&result)
	db.Disconnect(ctx)
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
	fmt.Fprintf(w, token)
	fmt.Println("Token generated: ", string(token))
	socketHandler(w, r)
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
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
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Received: %s", message)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}
	}
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/signup", Signup)
	http.HandleFunc("/signin", Signin)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func main() {
	fmt.Println("RUNNING ON PORT 8000...")
	handleRequests()
}

func connect() (*mongo.Client, context.Context) {
	clientOptions := options.Client()
	clientOptions.ApplyURI("mongodb://admin:admin@localhost:27017")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		fmt.Println("Error connecting to DB")
		panic(err)
	}
	return client, ctx
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
	if r.Header["Token"] == nil {
		fmt.Fprintf(w, "can not find token in header")
		return errors.New("Token error")
	}

	token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing")
		}
		return sampleSecretKey, nil
	})

	if token == nil {
		fmt.Fprintf(w, "invalid token")
		return errors.New("Token error")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Fprintf(w, "couldn't parse claims")
		return errors.New("Token error")
	}

	exp := claims["exp"].(float64)
	if int64(exp) < time.Now().Local().Unix() {
		fmt.Fprintf(w, "token expired")
		return errors.New("Token error")
	}

	return nil
}
