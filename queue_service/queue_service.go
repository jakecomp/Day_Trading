package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Transaction struct {
	Transaction_id string
	User_id        string
	Command        string
	Stock_id       string
	Stock_price    float32
	Cash_value     float32
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

var upgrader = websocket.Upgrader{}
var transactionQueue = make([]command, 0)

func enqueue(queue []command, element command) []command {
	queue = append(queue, element)
	//fmt.Println("Enqueued:", element)
	return queue
}

func dequeue(queue []command) (*command, []command) {

	if len(queue) == 0 {
		return nil, queue
	}

	element := &queue[0]
	if len(queue) == 1 {
		var tmp = []command{}
		return element, tmp
	}

	return element, queue[1:]
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate User
	fmt.Println("Endpoint Hit: WS")

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
	for {
		var transaction *command
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error during message reading:", err)
			break
		}

		//fmt.Println("Received: ", string(msg))

		var message Message
		// Message Format: {"Command" : "ENQUEUE" , "Data" : "Transaction{}" }
		err = json.Unmarshal(msg, &message)

		// Handle Enqueue and Dequeue
		if message.Command == "ENQUEUE" {
			transactionQueue = enqueue(transactionQueue, *message.Data)
		} else if message.Command == "DEQUEUE" {
			transaction, transactionQueue = dequeue(transactionQueue)
			// Empty check
			if transaction == nil {
				message.Command = "EMPTY"
				msg, err = json.Marshal(message)
				err = conn.WriteMessage(messageType, msg)
			} else {
				message.Command = "SUCCESS"
				message.Data = transaction
				msg, err = json.Marshal(message)
				err = conn.WriteMessage(messageType, msg)
			}
		} else {
			fmt.Println("Bad Request Format")
			fmt.Println("Request: ", message)
			err = conn.WriteMessage(websocket.TextMessage, []byte("Error: Bad Request Format"))
		}

		if err != nil {
			fmt.Println("Error during message writing:", err)
			break
		}
	}
}

func handleRequests() {
	http.HandleFunc("/ws", socketHandler)
	log.Fatal(http.ListenAndServe("10.9.0.7:8001", nil))
	// log.Fatal(http.ListenAndServe("localhost:8001", nil))
}

func main() {
	fmt.Println("Queue Service Starting... Port 8001")
	handleRequests()
}
