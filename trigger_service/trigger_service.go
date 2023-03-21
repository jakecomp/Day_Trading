package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

type quote struct {
	Stock string  `json:"stock"`
	Price float64 `json:"price"`
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

type Trigger struct {
	Amount  float64 `json:"amount"`
	Command string  `json:"command"`
	Stock   string  `json:"stock"`
	Price   float64 `json:"price"`
}

type TriggerKey struct {
	Stock   string
	Command string
}

type UserTriggers struct {
	User     string
	Triggers map[TriggerKey]Trigger
}

var userMap map[string]UserTriggers

var upgrader = websocket.Upgrader{}
var queueServiceConn *amqp.Channel
var queue amqp.Queue

func handleRequests() {
	http.HandleFunc("/ws", socketHandler)
	log.Fatal(http.ListenAndServe(":8004", nil))
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
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

func set_ammount(cmd string, user string, stock string, amount float64) {
	// Make Map Key (Stock, Command)
	triggerKey := TriggerKey{}
	triggerKey.Command = cmd
	triggerKey.Stock = stock

	// Check if key exists
	trigger, ok := userMap[user].Triggers[triggerKey]
	if ok {
		trigger.Amount = amount
	} else {
		trigger = Trigger{}
		trigger.Stock = stock
		trigger.Amount = amount
		trigger.Price = -1.0
	}

	userMap[user].Triggers[triggerKey] = trigger
}

func set_trigger(cmd string, user string, stock string, price float64) {
	// Make Map Key (Stock, Command)
	triggerKey := TriggerKey{}
	triggerKey.Command = cmd
	triggerKey.Stock = stock

	trigger, ok := userMap[user].Triggers[triggerKey]

	if ok {
		trigger.Price = price
		userMap[user].Triggers[triggerKey] = trigger
	}
}

func delete_key(cmd string, user string, stock string) {
	triggerKey := TriggerKey{stock, cmd}
	delete(userMap[user].Triggers, triggerKey)
}

// Grab quotes every second and compare to quote values
func socketReader(conn *websocket.Conn) {
	// Event Loop, Handle Comms in here
	cmd := &command{0, "NONE", []string{"TEST"}}
	for {
		// Recieve a trigger command
		_, message, err := conn.ReadMessage()
		err = json.Unmarshal(message, cmd)
		if err != nil {
			fmt.Println("Error during message reading:", err)
			break
		}

		if cmd.Command == "SET_BUY_AMOUNT" {
			amount, _ := strconv.ParseFloat(cmd.Args[2], 64)
			set_ammount("BUY", cmd.Args[0], cmd.Args[1], amount)
		} else if cmd.Command == "SET_BUY_TRIGGER" {
			price, _ := strconv.ParseFloat(cmd.Args[2], 64)
			set_trigger("BUY", cmd.Args[0], cmd.Args[1], price)
		} else if cmd.Command == "SET_SELL_AMOUNT" {
			amount, _ := strconv.ParseFloat(cmd.Args[2], 64)
			set_ammount("SELL", cmd.Args[0], cmd.Args[1], amount)
		} else if cmd.Command == "SET_SELL_TRIGGER" {
			price, _ := strconv.ParseFloat(cmd.Args[2], 64)
			set_trigger("SELL", cmd.Args[0], cmd.Args[1], price)
		} else if cmd.Command == "CANCEL_SET_BUY" {
			delete_key("BUY", cmd.Args[0], cmd.Args[1])
		} else if cmd.Command == "CANCEL_SET_SELL" {
			delete_key("SELL", cmd.Args[0], cmd.Args[1])
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func connectQueue(conn *amqp.Connection) (amqp.Queue, *amqp.Channel) {

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	// Declare a queue
	q, err := ch.QueueDeclare(
		"trigger", // Queue name
		false,     // Durable
		false,     // Delete when unused
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q, ch
}

func dial(url string) (*amqp.Connection, error) {
	for {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn, err
		}
		// Rabbitmq is slow to start so we might have to wait on it
		time.Sleep(time.Second * 3)
	}

}

func getQuote() *quote {
	// Get a quote
	resp, _ := http.Get("http://10.9.0.6:8002")
	quote := &quote{}
	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(quote)
	}
	return quote
}

func main() {
	log.SetOutput(ioutil.Discard)

	// Connect to RabbitMQ server
	time.Sleep(time.Second * 15)
	conn, err := dial("amqp://guest:guest@10.9.0.10:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	queue, queueServiceConn = connectQueue(conn)
	defer queueServiceConn.Close()

	userMap = make(map[string]UserTriggers)

	log.Println("RUNNING ON PORT 8004...")
	handleRequests()
}
