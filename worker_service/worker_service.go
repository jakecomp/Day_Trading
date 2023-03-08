package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type userid string
type Args []string
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
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Action       []string `xml:"action,attr" json:"action"`
}

type Command struct {
	Ticket  int
	Command string
	Args    Args
}

type Message struct {
	Command string
	Data    *Command
}

type Stock struct {
	name string
	cost float64
}

type User struct {
	Balance float64
	Stocks  map[string]*Stock
}

func NewUser() *User {
	return &User{
		Balance: float64(0),
		Stocks:  make(map[string]*Stock, 0),
	}
}

// Internal DB of user state
var users map[userid]*User

// Dispatch commands based on the command string given
func dispatch(cmd Command) (CMD, error) {
	log.Println("in dispatch command is ", cmd.Command, cmd.Args)
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

// Enqueue a new command to the queue server
// not used in this current implementation
func pushCommand(conn *websocket.Conn, t *Command) error {
	// Event Loop, Handle Comms in here
	fmt.Println("transaction: ", *t)

	message := &Message{"ENQUEUE", t}
	msg, _ := json.Marshal(*message)

	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}
	return err
}

// TODO avoid this blocking for to avoid unnecssecary blocking on main thread
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
			log.Println("Received: ", message)
			transaction := message.Data
			log.Println("Command: ", transaction)
			return message, nil
		} else if message.Command == "EMPTY" {
			// Empty, wait and try again
			time.Sleep(time.Millisecond * 500)
		} else {
			log.Println("Unknown Request")
			time.Sleep(time.Millisecond * 500)
		}

		if err != nil {
			return nil, err
		}
	}

}

// Used for logging anything related to a users account
func sendAccountLog(n *Notification, bal float64) {
	a := account_log{
		Username:     n.Userid,
		Funds:        fmt.Sprint(bal),
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

// User command logs
func sendUserLog(n *Notification) {
	var money string
	if users[userid(n.Userid)] == nil {
		money = "0"
	} else {
		money = fmt.Sprint(users[userid(n.Userid)].Balance)
	}
	u := user_log{
		Username:     n.Userid,
		Funds:        money,
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

// Not currently used
func sendSystemLog(n *Notification) {
	s := system_log{
		Username:     n.Userid,
		Funds:        fmt.Sprint(users[userid(n.Userid)].Balance),
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

// Logs incomming commands
func commandLogger(nch chan *Notification) {
	for {
		n := <-nch
		sendUserLog(n)
	}
}

// TODO clean this up there is so much repreated code
// TODO Publish errors as needed
func UserAccountManager(mb *MessageBus) {
	users = make(map[userid]*User)
	add := mb.SubscribeAll(notifyADD)
	buy := mb.SubscribeAll(notifyCOMMIT_BUY)
	sell := mb.SubscribeAll(notifyCOMMIT_SELL)

	for {
		select {
		case newMoney := <-add:
			uid := userid(newMoney.Userid)

			if users[uid] == nil {
				users[uid] = NewUser()
			}

			users[uid].Balance += *newMoney.Amount
			sendAccountLog(&newMoney, users[uid].Balance)
		case newMoney := <-sell:
			uid := userid(newMoney.Userid)

			if users[uid] == nil {
				users[uid] = NewUser()
			}

			if users[uid].Stocks[*newMoney.Stock] == nil {
				log.Fatalln("Trying to sell stock user doesn't own", *newMoney.Stock)
			}

			users[uid].Balance += *newMoney.Amount
			users[uid].Stocks[*newMoney.Stock] = nil
			sendAccountLog(&newMoney, users[uid].Balance)
		case newMoney := <-buy:
			uid := userid(newMoney.Userid)

			if users[uid] == nil {
				users[uid] = NewUser()
			}

			if users[uid].Balance < *newMoney.Amount {
				log.Fatalln("Negative balance is not allowed", *newMoney.Stock)
			}

			users[uid].Balance -= *newMoney.Amount
			users[uid].Stocks[*newMoney.Stock] = &Stock{
				name: *newMoney.Stock,
				cost: float64(0),
			}
			sendAccountLog(&newMoney, users[uid].Balance)
		}
	}
}

func main() {

	// Determine if we should use local host
	var host string
	if len(os.Args) > 1 {
		host = "localhost"
	} else {
		host = "10.9.0.7"
		// Disable logging by default
		log.SetOutput(ioutil.Discard)
	}
	queueServiceConn, _, _ := websocket.DefaultDialer.Dial("ws://"+host+":8001/ws?", nil)
	log.Println("Worker Service Starting...")

	// Message bus shared between commands
	mb := NewMessageBus()

	// intended for updating the DB when used
	ch := make(chan *Transaction)
	// Used for logging commands when recieved
	nch := make(chan *Notification)

	// Logs all transactions to user accounts
	go UserAccountManager(mb)
	// Logs all incoming commands
	go commandLogger(nch)

	for {
		select {
		case tra := <-ch:
			log.Println("pushing new transaction ", tra)
		default:
			t, err := getNextCommand(queueServiceConn)
			cmd, err := dispatch(*t.Data)
			if err == nil {
				// Log the new command
				go func() {
					n := cmd.Notify()
					nch <- &n
				}()
				// Execute this new command
				go Run(cmd, mb, ch)
				// Sleep is here to avoid blocking the queue server for too long
				time.Sleep(time.Millisecond * 10)
			} else {
				log.Println("ERROR:", err)
			}
		}

	}
}
