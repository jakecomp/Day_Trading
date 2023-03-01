package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	funcLookup := map[string]func(Command) (*report, error){
		"ADD":             add,
		"BUY":             buy,
		"COMMIT_BUY":      commit_buy,
		"CANCEL_BUY":      cancel_buy,
		"SELL":            sell,
		"COMMIT_SELL":     commit_sell,
		"CANCEL_SELL":     cancel_sell,
		"DUMPLOG":         dumplog,
		"DISPLAY_SUMMARY": display_summary,
	}
	funcLookup[cmd.Command](cmd)
}

// Purpose:
//
//	Add the given amount of money to the user's account
//
// Pre-conditions:
//
//	none
//
// Post-Conditions:
//
//	the user's account is increased by the amount of money specified
//
// Example:
//
//	ADD,jsmith,200.00
func add(cmd Command) (*report, error) {
	if len(cmd.Args) != 2 {
		return nil, errors.New("Wrong number of arguments for add")
	}
	// const user = a[0]
	// const amount = a[1]
	// TODO Add Money To DB
	// TODO Assert money in db went up by x amount
	return nil, errors.New("unfinished")
}

// Purpose:
//
//	Buy the dollar amount of the stock for the specified user at the
//	current price.
//
// Pre-conditions:
//
//	The user's account must be greater or equal to the amount of the
//	purchase.
//
// Post-Conditions:
//
//	The user is asked to confirm or cancel the transaction
//
// Example:
//
//	BUY,jsmith,ABC,200.00
func buy(cmd Command) (*report, error) {
	return nil, nil
}

// Purpose:
//
//	Commits the most recently executed BUY command
//
// Pre-conditions:
//
//	The user must have executed a BUY command within the previous 60
//	seconds
//
// Post-Conditions:
//
//		(a) the user's cash account is decreased by the amount user to
//		purchase the stock
//		(b) the user's account for the given stock is increased by the
//	    purchase amount
//
// Example:
//
//	COMMIT_BUY,jsmith
func commit_buy(cmd Command) (*report, error) {
	return nil, nil
}

// Purpose:
//
//	Cancels the most recently executed BUY Command
//
// Pre-conditions:
//
//	The user must have executed a BUY command within the previous 60
//	seconds
//
// Post-Conditions:
//
//	The last BUY command is canceled and any allocated system resources
//	are reset and released.
//
// Example:
//
//	CANCEL_BUY,jsmith
func cancel_buy(cmd Command) (*report, error) {
	return nil, nil
}

// Purpose:
//
//	Sell the specified dollar mount of the stock currently held by the
//	specified user at the current price.
//
// Pre-conditions:
//
//	The user's account for the given stock must be greater than or
//	equal to the amount being sold.
//
// Post-Conditions:
//
//	The user is asked to confirm or cancel the given transaction
//
// Example:
//
//	SELL,jsmith,ABC,100.00
func sell(cmd Command) (*report, error) {
	return nil, nil
}

// Purpose:
//
//	Commits the most recently executed SELL command
//
// Pre-conditions:
//
//	The user must have executed a SELL command within the previous 60
//	seconds
//
// Post-Conditions:
//
//	(a) the user's account for the given stock is decremented by the
//	sale amount
//	(b) the user's cash account is increased by the sell amount
//
// Example:
//
//	COMMIT_SELL,jsmith
func commit_sell(cmd Command) (*report, error) {
	return nil, nil
}

// Purpose:
//
//	Cancels the most recently executed SELL Command
//
// Pre-conditions:
//
//	The user must have executed a SELL command within the previous 60 seconds
//
// Post-Conditions:
//
//		The last SELL command is canceled and any allocated system
//	    resources are reset and released.
//
// Example:
//
//	CANCEL_SELL,jsmith
func cancel_sell(a Command) (*report, error) {
	return nil, nil
}

// Purpose:
//
//	Print out the history of the users transactions to the user specified file
//
// Pre-conditions:
//
//	none
//
// Post-Conditions:
//
//	The history of the user's transaction are written to the specified file.
//
// Example:
//
//	DUMPLOG,userid,filename
func dumplogUser(userid, filename) (*report, error) {
	return nil, nil
}

// Purpose:
//
//	Print out to the specified file the complete set of transactions that
//	have occurred in the system.
//
// Pre-conditions:
//
//	Can only be executed from the supervisor (root/administrator) account.
//
// Post-Conditions:
//
//	Places a complete log file of all transactions that have occurred in
//	the system into the file specified by filename
//
// Example:
//
//	DUMPLOG,out.dump
func dumplogAll(filename) (*report, error) {
	return nil, nil
}

// Purpose:
//
//	Provides a summary to the client of the given user's transaction
//	history and the current status of their accounts as well as any set
//	buy or sell triggers and their parameters
//
// Pre-conditions:
//
//	none
//
// Post-Conditions:
//
//	A summary of the given user's transaction history and the
//	current status of their accounts as well as any set buy or sell
//	triggers and their parameters is displayed to the user.
//
// Example:
//
//	DISPLAY_SUMMARY,userid
func display_summary(cmd Command) (*report, error) {
	return nil, nil
}

func dumplog(cmd Command) (*report, error) {
	a := cmd.Args
	switch len(a) {
	case 2:
		dumplogUser(userid(a[0]), filename(a[1]))
		break
	case 1:
		dumplogAll(filename(a[0]))
		break
	default:
		return nil, errors.New("Invalid number of arguments to DUMPLOG")
	}
	return nil, nil
}

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
	// TODO update to queue server IP based on docker-compose
	queueServiceConn, _, _ := websocket.DefaultDialer.Dial("ws://10.9.0.7:8001/ws?", nil)
	fmt.Println("Worker Service Starting...")
	socketReader(queueServiceConn)
}
