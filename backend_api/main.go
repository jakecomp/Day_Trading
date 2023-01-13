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

func signup(w http.ResponseWriter, r *http.Request) {
	db, ctx, cancel, dberr := connect("mongodb://localhost:27017")
	if dberr != nil {
		panic(dberr)
	}
	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Salt and hash the password using the bcrypt algorithm
	// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

	var doc = new(user_doc)
	doc.Username = creds.Username
	doc.Hash = string(hashedPassword)

	collection := db.Database(database).Collection("users")
	result, err := collection.InsertOne(ctx, doc)
	print(result)

	defer cancel()
	db.Disconnect(ctx)
}

func signin(w http.ResponseWriter, r *http.Request) {

}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/singin", signin)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func main() {
	handleRequests()
}

func connect(uri string) (*mongo.Client, context.Context,
	context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}
