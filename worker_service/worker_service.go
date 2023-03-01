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
type result string

func dispatch(cmd Command) {
	funcLookup := map[string]func(Args) error{
		"ADD":              add,
		"QUOTE":            quote,
		"BUY":              buy,
		"COMMIT_BUY":       commit_buy,
		"CANCEL_BUY":       cancel_buy,
		"SELL":             sell,
		"COMMIT_SELL":      commit_sell,
		"CANCEL_SELL":      cancel_sell,
		"SET_BUY_AMOUNT":   set_buy_amount,
		"CANCEL_SET_BUY":   cancel_set_buy,
		"SET_BUY_TRIGGER":  set_buy_trigger,
		"SET_SELL_AMOUNT":  set_sell_amount,
		"SET_SELL_TRIGGER": set_sell_trigger,
		"CANCEL_SET_SELL":  cancel_set_sell,
		"DUMPLOG":          dumplog,
		"DISPLAY_SUMMARY":  display_summary,
	}
	funcLookup[cmd.Command](cmd.Args)
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
func add(a Args) error {
	return nil
}

// Purpose:
//
//	Get the current quote for the stock for the specified user
//
// Pre-conditions:
//
//	none
//
// Post-Conditions:
//
//	the current price of the specified stock is displayed to the user
//
// Example:
//
//	QUOTE,jsmith,ABC
func quote(a Args) error {
	return nil
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
func buy(a Args) error {
	return nil
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
func commit_buy(a Args) error {
	return nil
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
func cancel_buy(a Args) error {
	return nil
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
func sell(a Args) error {
	return nil
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
func commit_sell(a Args) error {
	return nil
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
func cancel_sell(a Args) error {
	return nil
}

// Purpose:
//
//	Sets a defined amount of the given stock to buy when the current
//	stock price is less than or equal to the BUY_TRIGGER
//
// Pre-conditions:
//
//	The user's cash account must be greater than or equal to the BUY
//	amount at the time the transaction occurs
//
// Post-Conditions:
//
//	(a) a reserve account is created for the BUY transaction to hold the
//		specified amount in reserve for when the transaction is triggered
//	(b) the user's cash account is decremented by the specified amount
//	(c) when the trigger point is reached the user's stock account is
//		updated to reflect the BUY transaction.
//
// Example:
//
//	SET_BUY_AMOUNT,jsmith,ABC,50.00
func set_buy_amount(a Args) error {
	return nil
}

// Purpose:
//
//	Cancels a SET_BUY command issued for the given stock
//
// Pre-conditions:
//
//	The must have been a SET_BUY Command issued for the given stock
//	by the user
//
// Post-Conditions:
//
//	(a) All accounts are reset to the values they would have had had
//	the SET_BUY Command not been issued
//	(b) the BUY_TRIGGER for the given user and stock is also canceled.
//
// Example:
//
//	CANCEL_SET_BUY,jsmith,ABC
func cancel_set_buy(a Args) error {
	return nil
}

// Purpose:
//
//	Sets the trigger point base on the current stock price when any
//	SET_BUY will execute.
//
// Pre-conditions:
//
//	The user must have specified a SET_BUY_AMOUNT prior to setting a
//	SET_BUY_TRIGGER
//
// Post-Conditions:
//
//	The set of the user's buy triggers is updated to include the
//	specified trigger
//
// Example:
//
//	SET_BUY_TRIGGER,jsmith,ABC,20.00
func set_buy_trigger(a Args) error {
	return nil
}

// Purpose:
//
//	Sets a defined amount of the specified stock to sell when the
//	current stock price is equal or greater than the sell trigger point
//
// Pre-conditions:
//
//	The user must have the specified amount of stock in their account
//	for that stock.
//
// Post-Conditions:
//
//	A trigger is initialized for this username/stock symbol
//	combination, but is not complete until SET_SELL_TRIGGER is
//	executed.
//
// Example:
//
//	SET_SELL_AMOUNT,jsmith,ABC,550.50
func set_sell_amount(a Args) error {
	return nil
}

// Purpose:
//
//	Sets the stock price trigger point for executing any SET_SELL
//	triggers associated with the given stock and user
//
// Pre-conditions:
//
//	The user must have specified a SET_SELL_AMOUNT prior to setting a SET_SELL_TRIGGER
//
// Post-Conditions:
//
//	(a) a reserve account is created for the specified
//		amount of the given stock
//	(b) the user account for the given stock is
//		reduced by the max number of stocks that could be purchased and
//	(c) the set of the user's sell triggers is updated to include the
//		specified trigger.
//
// Example:
//
//	SET_SELL_TRIGGER, jsmith,ABC,120.00
func set_sell_trigger(a Args) error {
	return nil
}

// Purpose:
//
//	Cancels the SET_SELL associated with the given stock and user
//
// Pre-conditions:
//
//	The user must have had a previously set SET_SELL for the given stock
//
// Post-Conditions:
//
//	(a) The set of the user's sell triggers is updated to remove the
//		sell trigger associated with the specified stock
//	(b) all user account information is reset to the values they would
//		have been if the given SET_SELL command had not been issued
//
// Example:
//
//	CANCEL_SET_SELL,jsmith,ABC
func cancel_set_sell(Args) error {
	return nil
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
func dumplogUser(userid, filename) error {
	return nil
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
func dumplogAll(filename) error {
	return nil
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
func display_summary(Args) error {
	return nil
}

func dumplog(a Args) error {
	switch len(a) {
	case 2:
		dumplogUser(userid(a[0]), filename(a[1]))
		break
	case 1:
		dumplogAll(filename(a[0]))
		break
	default:
		return errors.New("Invalid number of arguments to DUMPLOG")
	}
	return nil
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
	queueServiceConn, _, _ := websocket.DefaultDialer.Dial("ws://localhost:8001/ws?", nil)
	fmt.Println("Worker Service Starting...")
	socketReader(queueServiceConn)
}
