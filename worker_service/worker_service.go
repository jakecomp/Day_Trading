package main

import (
	"encoding/json"
	"fmt"
	"time"

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

type Message struct {
	Command string
	Data    *Transaction
}

func socketReader(conn *websocket.Conn) {
	// Event Loop, Handle Comms in here
	transaction := &Transaction{"ID_1", "USERNAME", "BUY", "S", 24.5, 600.0}
	fmt.Println("transaction: ", *transaction)

	message := &Message{"ENQUEUE", transaction}
	msg, _ := json.Marshal(*message)

	fmt.Println("MSG: ", string(msg))
	err := conn.WriteMessage(websocket.TextMessage, msg)

	if err != nil {
		fmt.Println("Error during enqueue:", err)
	}

	for {
		// Attempt Dequeue
		message.Command = "DEQUEUE"
		message.Data = nil
		msg, err = json.Marshal(message)
		err = conn.WriteMessage(websocket.TextMessage, msg)

		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error during message reading:", err)
			break
		}

		fmt.Println("MSG: ", string(msg))

		err = json.Unmarshal(msg, message)

		fmt.Println("Received: ", message)
		if message.Command == "SUCCESS" {
			transaction = message.Data
			// DO STUFF WITH TRANSACTION
			fmt.Println("Transaction: ", transaction)
		} else if message.Command == "EMPTY" {
			// Empty, wait and try again
			time.Sleep(time.Millisecond * 5000)
		} else {
			fmt.Println("Unknown Request")
			time.Sleep(time.Millisecond * 5000)
		}

		if err != nil {
			fmt.Println("Error during message writing:", err)
			break
		}
	}
}

func main() {
	queueServiceConn, _, _ := websocket.DefaultDialer.Dial("ws://localhost:8001/ws?", nil)
	fmt.Println("Worker Service Starting...")
	socketReader(queueServiceConn)
}
