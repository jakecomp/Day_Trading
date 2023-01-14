package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var db *mongo.Client

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
	db, ctx, dberr := connect()
	if dberr != nil {
		fmt.Println("Error connecting to DB")
		panic(dberr)
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
	err = nil
	// Salt and hash the password using the bcrypt algorithm
	// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		fmt.Println("Error hashing: ", err)
		return
	}
	fmt.Println("Password has been hashed")
	var doc = new(user_doc)
	doc.Username = creds.Username
	doc.Hash = string(hashedPassword)
	err = nil
	collection := db.Database(database).Collection("users")
	result, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("Error Inserting to DB: ", err)
		return
	}
	fmt.Fprintf(w, "SIGNED YOU UP!")
	fmt.Println("signup user " + doc.Username + " with hash " + doc.Hash)

	print(result)

	defer db.Disconnect(ctx)
}

func Signin(w http.ResponseWriter, r *http.Request) {

}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/signup", Signup)
	http.HandleFunc("/singin", Signin)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func main() {
	handleRequests()
}

func connect() (*mongo.Client, context.Context, error) {
	clientOptions := options.Client()
	clientOptions.ApplyURI("mongodb://admin:admin@localhost:27017")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, clientOptions)
	return client, ctx, err
}
