package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
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

func dispatch(cmd Command) (CMD, error) {
	fmt.Println("in dispatch command is ", cmd.Command, cmd.Args)
	funcLookup := map[string]func(Command) (CMD, error){
		"ADD": func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[1], 64)
			return ADD{ticket: int64(cmd.Ticket), userId: cmd.Args[0], amount: a}, err
		},
		"BUY": func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return BUY{ticket: int64(cmd.Ticket), userId: cmd.Args[0], stock: cmd.Args[1], amount: a, cost: 0}, err
		},
		"COMMIT_BUY": func(cmd Command) (CMD, error) {
			return &COMMIT_BUY{ticket: int64(cmd.Ticket), userId: cmd.Args[0]}, nil
		},
		"CANCEL_BUY": func(cmd Command) (CMD, error) {
			return &CANCEL_BUY{ticket: int64(cmd.Ticket), userId: cmd.Args[0]}, nil
		},
		"SELL": func(cmd Command) (CMD, error) {
			a, _ := strconv.ParseFloat(cmd.Args[1], 64)
			return &SELL{ticket: int64(cmd.Ticket), userId: cmd.Args[0], stock: cmd.Args[1], amount: a, cost: 0}, nil
		},
		"COMMIT_SELL": func(cmd Command) (CMD, error) {
			return &COMMIT_SELL{ticket: int64(cmd.Ticket), userId: cmd.Args[0]}, nil
		},
		"CANCEL_SELL": func(cmd Command) (CMD, error) {
			return &CANCEL_SELL{ticket: int64(cmd.Ticket), userId: cmd.Args[0]}, nil
		},
	}
	f := funcLookup[cmd.Command]
	if f == nil {
		return nil, errors.New("Undefinined command" + cmd.Command)
	}
	return funcLookup[cmd.Command](cmd)
}

type Message struct {
	Command string
	Data    *Command
}

// func socketReader(conn *websocket.Conn) {
// 	// Event Loop, Handle Comms in here
// 	transaction := &Command{1, "BUY", Args{"USERNAME", "S", "24.5", "600.0"}}
// 	fmt.Println("transaction: ", *transaction)

// 	message := &Message{"ENQUEUE", transaction}
// 	msg, _ := json.Marshal(*message)

// 	fmt.Println("MSG: ", string(msg))
// 	err := conn.WriteMessage(websocket.TextMessage, msg)

// 	if err != nil {
// 		fmt.Println("Error during enqueue:", err)
// 	}

// 	for {
// 		// Attempt Dequeue
// 		message.Command = "DEQUEUE"
// 		message.Data = nil
// 		msg, err = json.Marshal(message)
// 		err = conn.WriteMessage(websocket.TextMessage, msg)

// 		_, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			fmt.Println("Error during message reading:", err)
// 			break
// 		}

// 		fmt.Println("MSG: ", string(msg))

// 		err = json.Unmarshal(msg, message)

// 		fmt.Println("Received: ", message)
// 		if message.Command == "SUCCESS" {
// 			transaction = message.Data
// 			// DO STUFF WITH TRANSACTION
// 			fmt.Println("Transaction: ", transaction)
// 		} else if message.Command == "EMPTY" {
// 			// Empty, wait and try again
// 			time.Sleep(time.Millisecond * 5000)
// 		} else {
// 			fmt.Println("Unknown Request")
// 			time.Sleep(time.Millisecond * 5000)
// 		}

// 		if err != nil {
// 			fmt.Println("Error during message writing:", err)
// 			break
// 		}
// 	}
// }

func pushCommand(conn *websocket.Conn, t *Command) error {
	// Event Loop, Handle Comms in here
	fmt.Println("transaction: ", *t)

	message := &Message{"ENQUEUE", t}
	msg, _ := json.Marshal(*message)

	// fmt.Println("MSG: ", string(msg))
	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}
	// t2, err := getNextCommand(conn)
	// if t2.Command != "SUCCESS" {
	// 	log.Fatal("Failed to push ", t2, err)
	// }
	return err
}

func getNextCommand(conn *websocket.Conn) (*Message, error) {
	for {
		// Attempt Dequeue
		message := &Message{"DEQUEUE", nil}
		msg, err := json.Marshal(message)
		err = conn.WriteMessage(websocket.TextMessage, msg)

		_, resp, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(resp, message)

		if message.Command == "SUCCESS" {
			fmt.Println("Received: ", message)
			transaction := message.Data
			fmt.Println("Command: ", transaction)
			return message, nil
		} else if message.Command == "EMPTY" {
			// Empty, wait and try again
			time.Sleep(time.Millisecond * 500)
		} else {
			fmt.Println("Unknown Request")
			time.Sleep(time.Millisecond * 500)
		}

		if err != nil {
			return nil, err
		}
	}

}

type user_log struct {
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Command      []string `xml:"command,attr" json:"command"`
}

type system_log struct {
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Command      []string `xml:"command,attr" json:"command"`
}

type account_log struct {
	Timestamp    int64    `xml:"timestamp"`
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Action       []string `xml:"action,attr" json:"action"`
}

func sendAccountLog(n *Notification) {
	a := account_log{
		Username:     n.Userid,
		Funds:        fmt.Sprintf("%f", n.Amount),
		Ticketnumber: int(n.Ticket),
		Action:       []string{n.Topic},
	}

	ulog, _ := json.Marshal(a)
	bodyReader := bytes.NewReader(ulog)
	_, err := http.Post("http://10.9.0.9:8004/accountlog", "application/json", bodyReader)
	if err != nil {
		log.Fatal(err)
	}
}

func sendUserLog(n *Notification) {
	u := user_log{
		Username:     n.Userid,
		Funds:        fmt.Sprintf("%f", n.Amount),
		Ticketnumber: int(n.Ticket),
		Command:      []string{n.Topic},
	}

	ulog, _ := json.Marshal(u)
	bodyReader := bytes.NewReader(ulog)
	_, err := http.Post("http://10.9.0.9:8004/userlog", "application/json", bodyReader)
	if err != nil {
		log.Fatal(err)
	}
}

func sendSystemLog(n *Notification) {
	s := system_log{
		Username:     n.Userid,
		Funds:        "0",
		Ticketnumber: int(n.Ticket),
		Command:      []string{n.Topic},
	}
	ulog, _ := json.Marshal(s)
	bodyReader := bytes.NewReader(ulog)
	_, err := http.Post("http://10.9.0.9:8004/systemlog", "application/json", bodyReader)
	if err != nil {
		log.Fatal(err)
	}
}
func merge[T any](cs ...<-chan T) <-chan T {
	out := make(chan T)

	for _, c := range cs {
		go func() {
			for v := range c {
				out <- v
			}
		}()
	}
	return out
}
func logger(nchan chan *Notification) {
	for {
		n := <-nchan
		sendUserLog(n)
	}

}

// XXX Not used since I wanted the commands to just be handled on dispatch
// func logger(mb *MessageBus) {
// 	add := mb.SubscribeAll(notifyADD)
// 	buy := mb.SubscribeAll(notifyBUY)
// 	commit_buy := mb.SubscribeAll(notifyCOMMIT_BUY)
// 	cancel_buy := mb.SubscribeAll(notifyCANCEL_BUY)
// 	sell := mb.SubscribeAll(notifySELL)
// 	commit_sell := mb.SubscribeAll(notifyCOMMIT_SELL)
// 	cancel_sell := mb.SubscribeAll(notifyCANCEL_SELL)

//		for {
//			select {
//			case a := <-add:
//				sendUserLog(&a)
//			case a := <-buy:
//				sendUserLog(&a)
//			case a := <-commit_buy:
//				sendUserLog(&a)
//			case a := <-cancel_buy:
//				sendUserLog(&a)
//			case a := <-sell:
//				sendUserLog(&a)
//			case a := <-commit_sell:
//				sendUserLog(&a)
//			case a := <-cancel_sell:
//				sendUserLog(&a)
//			}
//		}
//	}
func main() {
	// Determin if we should use local host
	var host string
	if len(os.Args) > 1 {
		host = "localhost"
	} else {
		host = "10.9.0.7"
	}
	queueServiceConn, _, _ := websocket.DefaultDialer.Dial("ws://"+host+":8001/ws?", nil)
	fmt.Println("Worker Service Starting...")
	// These are just here as an example of what the queue server
	// could be getting on the other end for the worker to preform
	// commands := []*Command{
	// 	{1, notifyADD, []string{"USERNAME", "50.5"}},
	// 	{2, notifyBUY, []string{"USERNAME", "XYZ", "24.5", "600.0"}},
	// 	{3, notifyCOMMIT_BUY, []string{"USERNAME"}},
	// }
	// // Enqueue Tasks
	// for _, c := range commands {
	// 	err := pushCommand(
	// 		queueServiceConn,
	// 		c,
	// 	)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	ch := make(chan *Transaction)
	mb := NewMessageBus()
	nch := make(chan *Notification)
	go logger(nch)
	for {
		select {
		case tra := <-ch:
			fmt.Println("pushing new transaction ", tra)
			// sendUserLog(tra)

		default:
			t, err := getNextCommand(queueServiceConn)
			cmd, err := dispatch(*t.Data)
			if err == nil {
				n := cmd.Notify()
				go func() {
					nch <- &n
				}()
				go Run(cmd, mb, ch)
				time.Sleep(time.Millisecond * 10)
			} else {
				log.Println("ERROR:", err)
			}
		}

	}
}
