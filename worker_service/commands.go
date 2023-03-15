package main

// TODO Add a way to lookup a users balance
// used for pre req for buy and sell
// used for post req for commit commands
// TODO Add a way to lookup a stock value
// used for selling and buying
// TODO assert timeframe for commit and cancel commands
import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Client

const database = "day_trading"

func connect() (*mongo.Client, context.Context) {
	clientOptions := options.Client()
	clientOptions.ApplyURI("mongodb://admin:admin@10.9.0.3:27017")
	// clientOptions.ApplyURI("mongodb://admin:admin@localhost:27017")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		fmt.Println("Error connecting to DB")
		panic(err)
	}
	return client, ctx
}

type user_doc struct {
	Username string
	Hash     string
	Balance  float32
}

type Transaction struct {
	Transaction_id int64
	User_id        string
	Command        string
	Stock_id       string
	Stock_price    float64
	Cash_value     float64
}

type CMD interface {
	Notify() Notification
	Prerequsite(*MessageBus) error
	Execute(ch chan *Transaction) error
	Postrequsite(*MessageBus) error
}

func Run(c CMD, m *MessageBus, tchan chan *Transaction) {
	log.Println("Executing prereq for ", reflect.TypeOf(c), c)
	err := c.Prerequsite(m)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Executed prereq for ", reflect.TypeOf(c), c)
	log.Println("Executing ", reflect.TypeOf(c), c)
	err = c.Execute(tchan)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Executed ", reflect.TypeOf(c), c)
	log.Println("Sending notification for ", reflect.TypeOf(c), c)
	go func() {
		n := c.Notify()
		m.Publish(n.Topic, n)
	}()
	log.Println("Notification sent for ", reflect.TypeOf(c), c)
	log.Println("Executing Postreq for ", reflect.TypeOf(c), c)
	err = c.Postrequsite(m)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Executed Postreq for ", reflect.TypeOf(c), c)
}

const (
	notifyADD                = "ADD"
	notifyBUY                = "BUY"
	notifySELL               = "SELL"
	notifyCOMMIT_BUY         = "COMMIT_BUY"
	notifyCOMMIT_SELL        = "COMMIT_SELL"
	notifyCANCEL_BUY         = "CANCEL_BUY"
	notifyCANCEL_SELL        = "CANCEL_SELL"
	notifyCANCEL_SET_SELL    = "CANCEL_SET_SELL"
	notifySET_SELL_TRIGGER   = "SET_SELL_TRIGGER"
	notifySET_SELL_AMOUNT    = "SET_SELL_AMOUNT"
	notifySET_BUY_TRIGGER    = "SET_BUY_TRIGGER"
	notifyCANCEL_BUY_TRIGGER = "CANCEL_BUY_TRIGGER"
	notifySET_BUY_AMOUNT     = "SET_BUY_AMOUNT"
	notifySTOCK_PRICE        = "STOCK_PRICE"
)

func read_db(username string, add_command bool) (user_collection *user_doc) {

	db, ctx := connect()

	var err error
	var result user_doc
	err = db.Database(database).Collection("users").FindOne(ctx, bson.D{{"username", username}}).Decode(&result)

	if err != nil {

		if err.Error() == "mongo: no documents in result" && add_command {

			var new_doc = new(user_doc)
			new_doc.Username = username
			new_doc.Hash = "unsecure_this_user_never_made_account_via_backend"
			new_doc.Balance = 0

			collection := db.Database(database).Collection("users")
			_, err = collection.InsertOne(context.TODO(), new_doc)

			if err != nil {
				fmt.Println("Error inserting into db: ", err)
				panic(err)
			}

			db.Disconnect(ctx)
			return new_doc

		} else {

			fmt.Println("Error search for record: ", err)
			panic(err)
		}

	}

	db.Disconnect(ctx)
	return &result
}

func update_db(new_doc *user_doc) {

	db, ctx := connect()

	var err error

	collection := db.Database(database).Collection("users")

	selected_user := bson.M{"username": new_doc.Username}
	updated_user := bson.M{"$set": bson.M{"balance": new_doc.Balance}}
	_, err = collection.UpdateOne(context.TODO(), selected_user, updated_user)
	db.Disconnect(ctx)

	if err != nil {
		fmt.Println("Error inserting into db: ", err)
		panic(err)
	}
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
type ADD struct {
	ticket int64
	userId string
	amount float64
}

func (a ADD) Notify() Notification {
	return Notification{
		Topic:     notifyADD,
		Timestamp: time.Now(),
		Ticket:    a.ticket,
		Userid:    a.userId,
		Stock:     nil,
		Amount:    &a.amount,
	}
}

// lookup user balance
// if invalid balance return an error describing this
func (a ADD) Prerequsite(mb *MessageBus) error {
	return nil
}
func (a ADD) Execute(ch chan *Transaction) error {
	ch <- &Transaction{
		Transaction_id: a.ticket,
		User_id:        a.userId,
		Command:        notifyADD,
		Cash_value:     a.amount,
		Stock_id:       "",
		Stock_price:    0,
	}

	fmt.Println(a.userId)
	var user_account user_doc = *read_db(a.userId, true)
	user_account.Balance = user_account.Balance + float32(a.amount)

	if user_account.Balance == 0 {
		fmt.Println("ADD ERROR! No update to balance")
	}
	update_db(&user_account)
	return nil
}

func (a ADD) Postrequsite(mb *MessageBus) error {
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
type BUY struct {
	ticket int64
	userId string
	stock  string
	amount float64
}

// lookup user balance
// if invalid balance return an error describing this
func (b BUY) Prerequsite(mb *MessageBus) error {
	// ch := mb.Subscribe(notifyADD, userid(b.userId))
	// for n := range ch {
	// 	if n.Userid == b.userId {
	// 		if *n.Amount < b.amount {
	// 			return errors.New("Not enough money for this stock")
	// 		} else if n.Ticket < b.ticket {
	return nil
	// 		}
	// 	}
	// }
	// return errors.New("Balance Channel Prematurely Closed")

}

// Removed since nothing should actually happen until the commit or cancel
func (b BUY) Execute(ch chan *Transaction) error {
	return nil
}

func (b BUY) Notify() Notification {
	return Notification{
		Topic:     notifyBUY,
		Timestamp: time.Now(),
		Ticket:    b.ticket,
		Userid:    b.userId,
		Stock:     &b.stock,
		Amount:    &b.amount,
	}
}

func (b BUY) Postrequsite(mb *MessageBus) error {
	commitChan := mb.Subscribe(notifyCOMMIT_BUY, userid(b.userId))
	cancelChan := mb.Subscribe(notifyCANCEL_BUY, userid(b.userId))

	select {
	case n := <-commitChan:
		if n.Userid == b.userId && n.Ticket > b.ticket {
			return nil
		}

	case n := <-cancelChan:
		if n.Userid == b.userId && n.Ticket > b.ticket {
			return nil
		}
	}
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
//	(a) the user's cash account is decreased by the amount user to
//	    purchase the stock
//	(b) the user's account for the given stock is increased by the
//	    purchase amount
//
// Example:
//
//	COMMIT_BUY,jsmith

type COMMIT_BUY struct {
	ticket int64
	userId string
	buy    Notification
}

func (b *COMMIT_BUY) Prerequsite(mb *MessageBus) error {
	ch := mb.Subscribe(notifyBUY, userid(b.userId))

	for n := range ch {
		if n.Userid == b.userId && n.Ticket < b.ticket {
			b.buy = n
			return nil
		}
	}
	return errors.New("Balance Channel Prematurely Closed")

}

func (b COMMIT_BUY) Execute(ch chan *Transaction) error {
	ch <- &Transaction{
		Transaction_id: b.ticket,
		User_id:        b.userId,
		Command:        notifyCOMMIT_BUY,
		Stock_id:       *b.buy.Stock,
		Cash_value:     *b.buy.Amount,
		Stock_price:    0,
	}
	return nil
}

func (b COMMIT_BUY) Notify() Notification {
	return Notification{
		Topic:     notifyCOMMIT_BUY,
		Timestamp: time.Now(),
		Ticket:    b.ticket,
		Userid:    b.userId,
		Stock:     b.buy.Stock,
		Amount:    b.buy.Amount,
	}
}

func (b COMMIT_BUY) Postrequsite(mb *MessageBus) error {
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
type CANCEL_BUY struct {
	ticket int64
	userId string
	buy    Notification
}

func (b *CANCEL_BUY) Prerequsite(mb *MessageBus) error {
	// ch := mb.Subscribe(notifyBUY, userid(b.userId))

	// for n := range ch {
	// 	if n.Userid == b.userId && n.Ticket < b.ticket {
	// 		b.buy = n
	// 		return nil
	// 	}
	// }
	// return errors.New("Balance Channel Prematurely Closed")
	return nil
}

func (b CANCEL_BUY) Execute(ch chan *Transaction) error {
	return nil
}

func (b CANCEL_BUY) Notify() Notification {
	return Notification{
		Topic:     notifyCANCEL_BUY,
		Timestamp: time.Now(),
		Ticket:    b.ticket,
		Userid:    b.userId,
		Stock:     b.buy.Stock,
		Amount:    b.buy.Amount,
	}
}

func (b CANCEL_BUY) Postrequsite(mb *MessageBus) error {
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
type SELL struct {
	ticket int64
	userId string
	stock  string
	amount float64
}

func (s SELL) Prerequsite(mb *MessageBus) error {
	// // Wait for the user to buy this stock
	// ch := mb.Subscribe(notifyCOMMIT_BUY, userid(s.userId))

	// for n := range ch {
	// 	if n.Userid == s.userId {
	// 		// if *n.Amount > s.amount {
	// 		// 	return errors.New(fmt.Sprintf("Don't have %f of %s", s.amount, s.stock))
	// 		// } else {
	// 		return nil
	// 		// }
	// 	}
	// }
	// return errors.New("Balance Channel Prematurely Closed")
	return nil
}
func (b SELL) Execute(ch chan *Transaction) error {
	return nil
}

func (s SELL) Notify() Notification {
	return Notification{
		Topic:     notifySELL,
		Timestamp: time.Now(),
		Ticket:    s.ticket,
		Userid:    s.userId,
		Stock:     &s.stock,
		Amount:    &s.amount,
	}
}

func (s SELL) Postrequsite(mb *MessageBus) error {
	// TODO determine how to lookup user balance
	commitChan := mb.Subscribe(notifyCOMMIT_SELL, userid(s.userId))
	cancelChan := mb.Subscribe(notifyCANCEL_SELL, userid(s.userId))
	select {
	case n := <-commitChan:
		if n.Userid == s.userId && n.Ticket > s.ticket {
			return nil
		}

	case n := <-cancelChan:
		if n.Userid == s.userId && n.Ticket > s.ticket {
			return nil
		}
	}
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
type COMMIT_SELL struct {
	ticket int64
	userId string
	sell   Notification
}

func (s *COMMIT_SELL) Prerequsite(mb *MessageBus) error {
	ch := mb.Subscribe(notifySELL, userid(s.userId))

	for n := range ch {
		if n.Userid == s.userId && n.Ticket < s.ticket {
			s.sell = n
			return nil
		}
	}
	return errors.New("Balance Channel Prematurely Closed")

}

func (s COMMIT_SELL) Execute(ch chan *Transaction) error {
	ch <- &Transaction{
		Transaction_id: s.ticket,
		User_id:        s.userId,
		Command:        notifyCOMMIT_SELL,
		Stock_id:       *s.sell.Stock,
		Stock_price:    *s.sell.Amount,
		Cash_value:     0,
	}
	return nil
}

func (s COMMIT_SELL) Notify() Notification {
	return Notification{
		Topic:     notifyCOMMIT_SELL,
		Timestamp: time.Now(),
		Ticket:    s.ticket,
		Userid:    s.userId,
		Stock:     s.sell.Stock,
		Amount:    s.sell.Amount,
	}
}

func (b COMMIT_SELL) Postrequsite(mb *MessageBus) error {
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
//	The last SELL command is canceled and any allocated system
//	resources are reset and released.
//
// Example:
//
//	CANCEL_SELL,jsmith
type CANCEL_SELL struct {
	ticket int64
	userId string
	sell   Notification
}

func (s *CANCEL_SELL) Prerequsite(mb *MessageBus) error {
	// ch := mb.Subscribe(notifySELL, userid(s.userId))

	// for n := range ch {
	// 	if n.Userid == s.userId && n.Ticket < s.ticket {
	// 		s.sell = n
	// 		return nil
	// 	}
	// }
	// return errors.New("Balance Channel Prematurely Closed")
	return nil

}

func (s CANCEL_SELL) Execute(ch chan *Transaction) error {
	return nil
}

func (s CANCEL_SELL) Notify() Notification {
	return Notification{
		Topic:     notifyCANCEL_SELL,
		Timestamp: time.Now(),
		Ticket:    s.ticket,
		Userid:    s.userId,
		Stock:     s.sell.Stock,
		Amount:    s.sell.Amount,
	}
}

func (b CANCEL_SELL) Postrequsite(mb *MessageBus) error {
	return nil
}

// Purpose:
//
// Cancels the SET_SELL associated with the given stock and user
// Pre-conditions:
//
// The user must have had a previously set SET_SELL for the given stock
// Post-Conditions:
//
// (a) The set of the user's sell triggers is updated to remove the
//
//	sell trigger associated with the specified stock
//
// (b) all user account information is reset to the values they would
//
//	have been if the given SET_SELL command had not been issued
//
// Example:
//
// CANCEL_SET_SELL,jsmith,ABC
type CANCEL_SET_SELL struct {
	ticket int64
	userId string
	sell   Notification
}

func (c CANCEL_SET_SELL) Prerequsite(mb *MessageBus) error {
	return nil
}

func (c CANCEL_SET_SELL) Notify() Notification {
	return Notification{
		Topic:     notifyCANCEL_SET_SELL,
		Timestamp: time.Now(),
		Ticket:    c.ticket,
		Userid:    c.userId,
		Stock:     nil,
		Amount:    nil,
	}
}
func (c CANCEL_SET_SELL) Execute(ch chan *Transaction) error {
	return nil
}
func (c CANCEL_SET_SELL) Postrequsite(mb *MessageBus) error {
	return nil
}

// Purpose:
//
// Sets the stock price trigger point for executing any SET_SELL
// triggers associated with the given stock and user
// Pre-conditions:
//
// The user must have specified a SET_SELL_AMOUNT prior to setting a SET_SELL_TRIGGER
// Post-Conditions:
//
// (a) a reserve account is created for the specified
//
//	amount of the given stock
//
// (b) the user account for the given stock is
//
//	reduced by the max number of stocks that could be purchased and
//
// (c) the set of the user's sell triggers is updated to include the
//
//	specified trigger.
//
// Example:
//
// SET_SELL_TRIGGER, jsmith,ABC,120.00
type SET_SELL_TRIGGER struct {
	ticket int64
	userId string
	sell   Notification
}

func (c SET_SELL_TRIGGER) Prerequsite(mb *MessageBus) error {
	return nil
}

func (c SET_SELL_TRIGGER) Notify() Notification {
	return Notification{
		Topic:     notifySET_SELL_TRIGGER,
		Timestamp: time.Now(),
		Ticket:    c.ticket,
		Userid:    c.userId,
		Stock:     nil,
		Amount:    nil,
	}
}
func (c SET_SELL_TRIGGER) Execute(ch chan *Transaction) error {
	return nil
}
func (c SET_SELL_TRIGGER) Postrequsite(mb *MessageBus) error {
	return nil
}

// Purpose:
//
// Sets a defined amount of the specified stock to sell when the
// current stock price is equal or greater than the sell trigger point
// Pre-conditions:
//
// The user must have the specified amount of stock in their account
// for that stock.
// Post-Conditions:
//
// A trigger is initialized for this username/stock symbol
// combination, but is not complete until SET_SELL_TRIGGER is
// executed.
// Example:
//
// SET_SELL_AMOUNT,jsmith,ABC,550.50
// TODO fully implement
type SET_SELL_AMOUNT struct {
	ticket int64
	userId string
	sell   Notification
}

func (c SET_SELL_AMOUNT) Prerequsite(mb *MessageBus) error {
	return nil
}

func (c SET_SELL_AMOUNT) Notify() Notification {
	return Notification{
		Topic:     notifySET_SELL_AMOUNT,
		Timestamp: time.Now(),
		Ticket:    c.ticket,
		Userid:    c.userId,
		Stock:     nil,
		Amount:    nil,
	}
}
func (c SET_SELL_AMOUNT) Execute(ch chan *Transaction) error {
	return nil
}
func (c SET_SELL_AMOUNT) Postrequsite(mb *MessageBus) error {
	return nil
}

// Purpose:
//
// Sets the trigger point base on the current stock price when any
// SET_BUY will execute.
// Pre-conditions:
//
// The user must have specified a SET_BUY_AMOUNT prior to setting a
// SET_BUY_TRIGGER
// Post-Conditions:
//
// The set of the user's buy triggers is updated to include the
// specified trigger
// Example:
//
// SET_BUY_TRIGGER,jsmith,ABC,20.00
// TODO fully implement
type SET_BUY_TRIGGER struct {
	ticket int64
	userId string
	sell   Notification
}

func (c SET_BUY_TRIGGER) Prerequsite(mb *MessageBus) error {
	return nil
}

func (c SET_BUY_TRIGGER) Notify() Notification {
	return Notification{
		Topic:     notifySET_BUY_TRIGGER,
		Timestamp: time.Now(),
		Ticket:    c.ticket,
		Userid:    c.userId,
		Stock:     nil,
		Amount:    nil,
	}
}
func (c SET_BUY_TRIGGER) Execute(ch chan *Transaction) error {
	return nil
}
func (c SET_BUY_TRIGGER) Postrequsite(mb *MessageBus) error {
	return nil
}

// Purpose:
//
// Cancels a SET_BUY command issued for the given stock
// Pre-conditions:
//
// The must have been a SET_BUY Command issued for the given stock
// by the user
// Post-Conditions:
//
// (a) All accounts are reset to the values they would have had had
// the SET_BUY Command not been issued
// (b) the BUY_TRIGGER for the given user and stock is also canceled.
// Example:
//
// CANCEL_SET_BUY,jsmith,ABC
// TODO fully implement
type CANCEL_BUY_TRIGGER struct {
	ticket int64
	userId string
	sell   Notification
}

func (c CANCEL_BUY_TRIGGER) Prerequsite(mb *MessageBus) error {
	return nil
}

func (c CANCEL_BUY_TRIGGER) Notify() Notification {
	return Notification{
		Topic:     notifyCANCEL_BUY_TRIGGER,
		Timestamp: time.Now(),
		Ticket:    c.ticket,
		Userid:    c.userId,
		Stock:     nil,
		Amount:    nil,
	}
}
func (c CANCEL_BUY_TRIGGER) Execute(ch chan *Transaction) error {
	return nil
}
func (c CANCEL_BUY_TRIGGER) Postrequsite(mb *MessageBus) error {
	return nil
}

// Purpose:
//
// Sets a defined amount of the given stock to buy when the current
// stock price is less than or equal to the BUY_TRIGGER
// Pre-conditions:
//
// The user's cash account must be greater than or equal to the BUY
// amount at the time the transaction occurs
// Post-Conditions:
//
// (a) a reserve account is created for the BUY transaction to hold the
// 	specified amount in reserve for when the transaction is triggered
// (b) the user's cash account is decremented by the specified amount
// (c) when the trigger point is reached the user's stock account is
// 	updated to reflect the BUY transaction.
// Example:
//
// SET_BUY_AMOUNT,jsmith,ABC,50.00

// TODO fully implement
type SET_BUY_AMOUNT struct {
	ticket int64
	userId string
	sell   Notification
}

func (c SET_BUY_AMOUNT) Prerequsite(mb *MessageBus) error {
	return nil
}

func (c SET_BUY_AMOUNT) Notify() Notification {
	return Notification{
		Topic:     notifySET_BUY_AMOUNT,
		Timestamp: time.Now(),
		Ticket:    c.ticket,
		Userid:    c.userId,
		Stock:     nil,
		Amount:    nil,
	}
}
func (c SET_BUY_AMOUNT) Execute(ch chan *Transaction) error {
	return nil
}
func (c SET_BUY_AMOUNT) Postrequsite(mb *MessageBus) error {
	return nil
}
