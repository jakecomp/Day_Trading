package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type Command struct {
	Ticket  int
	Command string
	Args    Args
}

type userid string
type amount float32
type StockSymbol string
type filename string
type Args []string
type report string

func dispatch(cmd Command) {
	funcLookup := map[string]func(Command) CMD{
		"ADD": func(cmd Command) CMD {
			a, _ := strconv.ParseFloat(cmd.Args[1], 64)
			return ADD{userId: cmd.Args[0], amount: a}
		},
		"BUY": func(cmd Command) CMD {
			a, _ := strconv.ParseFloat(cmd.Args[1], 64)
			return BUY{userId: cmd.Args[0], stock: cmd.Args[1], amount: a, cost: 0}
		},
		"COMMIT_BUY": func(cmd Command) CMD {
			return &COMMIT_BUY{userId: cmd.Args[0]}
		},
		"CANCEL_BUY": func(cmd Command) CMD {
			return &CANCEL_BUY{userId: cmd.Args[0]}
		},
		"SELL": func(cmd Command) CMD {
			a, _ := strconv.ParseFloat(cmd.Args[1], 64)
			return &SELL{userId: cmd.Args[0], stock: cmd.Args[1], amount: a, cost: 0}
		},
		"COMMIT_SELL": func(cmd Command) CMD {
			return &COMMIT_SELL{userId: cmd.Args[0]}
		},
		"CANCEL_SELL": func(cmd Command) CMD {
			return &CANCEL_SELL{userId: cmd.Args[0]}
		},
	}
	funcLookup[cmd.Command](cmd)
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

// func main() {
// 	// TODO update to queue server IP based on docker-compose
// 	queueServiceConn, _, _ := websocket.DefaultDialer.Dial("ws://10.9.0.7:8001/ws?", nil)
// 	fmt.Println("Worker Service Starting...")
// 	socketReader(queueServiceConn)
// }
