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
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9"
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Hour)
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
	Stonks   map[string]float64
}

type quote struct {
	Stock string  `json:"stock"`
	Price float64 `json:"price"`
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
	UserId() string
	Notify() Notification
	Prerequsite(*redis.Client) error
	// Postrequsite(*MessageBus) error
}

type StockPricer struct {
	prices map[string]Stock
	lock   sync.RWMutex
}

func (s *StockPricer) lookupPrice(stock string, ticket int64) Stock {
	// We use RLock here to allow for concurrent reads so long as
	// Lock has not been used. This prevents reading during writs
	s.lock.RLock()
	p, ok := s.prices[stock]
	s.lock.RUnlock()

	// Fallback if we still don't have a stock price
	if !ok {
		p = getQuote(stock)

		s.lock.Lock()
		s.prices[stock] = p
		s.lock.Unlock()

		sendDebugLog(int64(ticket),
			fmt.Sprint("Had to look up stock manually for",
				stock, "and got \n",
				p.Name, "for ", p.Price))
	}
	return p

}

func UserAccountManager(mb *MessageBus, notification Notification, db *mongo.Client, ctx context.Context, s *StockPricer) {
	// Map storing all the currently known stock prices
	var err error
	last_ticket := -1
	switch notification.Topic {
	// case t2price := <-stockPrice:
	// 	stockPrices[*t2price.Stock] = Stock{*t2price.Stock, *t2price.Amount}
	case notifyADD:
		err = addMoney(notification, db, &ctx)
		last_ticket = int(notification.Ticket)
	case notifyCOMMIT_SELL:
		p := s.lookupPrice(*notification.Stock, notification.Ticket)
		err = sellStock(p.Price, notification, db, &ctx)
		last_ticket = int(notification.Ticket)
	case notifyCOMMIT_BUY:
		p := s.lookupPrice(*notification.Stock, notification.Ticket)
		// Fallback if we still don't have a stock prices
		err = buyStock(p.Price, notification, db, &ctx)
		last_ticket = int(notification.Ticket)
	case notifyFORCE_SELL:
		// Fallback if we still don't have a stock price
		p := s.lookupPrice(*notification.Stock, notification.Ticket)
		err = sellStock(p.Price, notification, db, &ctx)
		last_ticket = int(notification.Ticket)
	case notifyFORCE_BUY:
		p := s.lookupPrice(*notification.Stock, notification.Ticket)
		// Fallback if we still don't have a stock price
		err = buyStock(p.Price, notification, db, &ctx)
		last_ticket = int(notification.Ticket)
		// default:

	}

	if err != nil {
		sendErrorLog(int64(last_ticket), fmt.Sprint("ERROR:", err))
	}
}

// timout for a user
// perticket
// last active timestamp
func Run(task CMD, m *MessageBus, rdb *redis.Client, db *mongo.Client, ctx context.Context, s *StockPricer) {
	log.Println("Executing prereq for ", reflect.TypeOf(task), task)
	err := task.Prerequsite(rdb)
	if err != nil {
		sendErrorLog(int64(task.Notify().Ticket), fmt.Sprint("ERROR:", err))
		return
	}
	log.Println("Executed ", reflect.TypeOf(task), task)
	log.Println("Sending notification for ", reflect.TypeOf(task), task)

	n := task.Notify()
	UserAccountManager(m, n, db, ctx, s)
	m.Publish(n.Topic, n)

}

const (
	notifyADD         = "ADD"
	notifyBUY         = "BUY"
	notifySELL        = "SELL"
	notifyCOMMIT_BUY  = "COMMIT_BUY"
	notifyCOMMIT_SELL = "COMMIT_SELL"
	notifyCANCEL_BUY  = "CANCEL_BUY"
	notifyCANCEL_SELL = "CANCEL_SELL"
	notifyFORCE_BUY   = "FORCE_BUY"
	notifyFORCE_SELL  = "FORCE_SELL"
	notifySTOCK_PRICE = "STOCK_PRICE"
)

func read_db(username string, add_command bool, db *mongo.Client, ctx context.Context) (user_collection *user_doc, err error) {

	if db == nil {
		db, ctx = connect()
	}
	var result user_doc
	err = db.Database(database).Collection("users").FindOne(ctx, bson.D{{"username", username}}).Decode(&result)

	if err != nil {
		if err.Error() == "mongo: no documents in result" && add_command {

			var new_doc = new(user_doc)
			new_doc.Username = username
			new_doc.Hash = "unsecure_this_user_never_made_account_via_backend"
			new_doc.Balance = 0
			new_doc.Stonks = make(map[string]float64)

			collection := db.Database(database).Collection("users")
			_, err = collection.InsertOne(context.TODO(), new_doc)

			if err != nil {
				fmt.Println("Error adding user to db: ", err)
				panic(err)
			}

			//defer db.Disconnect(ctx)
			return new_doc, nil

		} else {
			return nil, (err)
		}

	}
	return &result, nil
}

func update_db(new_doc *user_doc, db *mongo.Client, ctx context.Context) {
	if db == nil {
		db, ctx = connect()
	}

	collection := db.Database(database).Collection("users")

	selected_user := bson.M{"username": new_doc.Username}
	updated_user := bson.M{"$set": bson.M{"balance": new_doc.Balance, "stonks": new_doc.Stonks}}
	_, err := collection.UpdateOne(context.TODO(), selected_user, updated_user)

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

func (a ADD) UserId() string {
	return a.userId
}
func (a BUY) UserId() string {
	return a.userId
}
func (a SELL) UserId() string {
	return a.userId
}
func (a COMMIT_BUY) UserId() string {
	return a.userId
}
func (a COMMIT_SELL) UserId() string {
	return a.userId
}
func (a CANCEL_SELL) UserId() string {
	return a.userId
}
func (a CANCEL_BUY) UserId() string {
	return a.userId
}
func (a FORCE_BUY) UserId() string {
	return a.userId
}

func (a FORCE_SELL) UserId() string {
	return a.userId
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
func (a ADD) Prerequsite(rdb *redis.Client) error {
	if a.amount < 0 {
		return errors.New(fmt.Sprint("Attempt to add negative value ", a.amount, " to user ", a.userId))
	}
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
func (b BUY) Prerequsite(rdb *redis.Client) error {
	if b.amount < 0 {
		return errors.New(fmt.Sprint("Attempt to buy a negative amount of ", b.stock, " that amount being ", b.amount, " for user ", b.userId))
	}
	n := b.Notify()
	n.Pending(rdb)
	log.Println("set pending buy for ", b)
	return nil
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

func (b *COMMIT_BUY) Prerequsite(rdb *redis.Client) error {
	n, err := lastPending(b.userId, notifyBUY, rdb)
	if err != nil {
		log.Println(b, "failed to commit purchase")
		return err
	}
	if n != nil && n.Userid == b.userId && n.Ticket < b.ticket {
		log.Println(b, "commited purchase of ", *b)
		b.buy = *n
		return nil
	}

	return errors.New(fmt.Sprintln("No lingering buy found for ", b, "with n ", n))

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

func (b *CANCEL_BUY) Prerequsite(rdb *redis.Client) error {
	n, err := lastPending(b.userId, notifyBUY, rdb)
	if err != nil {
		log.Println(b, "failed to canceled purchase")
		return err
	}
	if n != nil && n.Userid == b.userId && n.Ticket < b.ticket {
		log.Println(b, "canceled the purchase of ", *n)
		sendDebugLog(b.ticket, fmt.Sprintln("Cancelling purchase of ", *n, " with ", b))

		return nil
	}

	return errors.New(fmt.Sprintln("No lingering buy found for ", b, "with n ", n))
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

func (s SELL) Prerequsite(rdb *redis.Client) error {
	if s.amount < 0 {
		return errors.New(fmt.Sprint("Attempt to sell a negative amount of ", s.stock, " that amount being ", s.amount, " for user ", s.userId))
	}
	n := s.Notify()
	n.Pending(rdb)
	log.Println("set pending sell for ", s)

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

func (s *COMMIT_SELL) Prerequsite(rdb *redis.Client) error {
	n, err := lastPending(s.userId, notifySELL, rdb)
	if err != nil {
		return err
	}

	if n != nil && n.Userid == s.userId && n.Ticket < s.ticket {
		log.Println(s, "commited the sale of ", *n)
		s.sell = *n
		return nil
	}

	return errors.New(fmt.Sprintln("No lingering sells found for ", s, "with n ", n))

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

func (s *CANCEL_SELL) Prerequsite(rdb *redis.Client) error {
	n, err := lastPending(s.userId, notifySELL, rdb)
	if err != nil {
		log.Println(s, "failed to find a sale to cancel")
		return err
	}
	if n != nil && n.Userid == s.userId && n.Ticket < s.ticket {
		log.Println(s, "cancelled the sale of ", *n)
		s.sell = *n
		sendDebugLog(s.ticket, fmt.Sprintln("Cancelling sale of ", *n, " with ", s))
		return nil
	}

	return errors.New(fmt.Sprintln("No lingering sells found for ", s, "with n ", n))
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
//	Commit Buy the dollar amount of the stock for the specified user at the
//	current price.
//
// Pre-conditions:
//
//	The user's account must be greater or equal to the amount of the
//	purchase.
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
//	FORCE_BUY,jsmith,ABC,200.00
type FORCE_BUY struct {
	ticket int64
	userId string
	stock  string
	amount float64
}

func (b *FORCE_BUY) Prerequsite(rdb *redis.Client) error {
	return nil
}

func (b FORCE_BUY) Execute(ch chan *Transaction) error {
	ch <- &Transaction{
		Transaction_id: b.ticket,
		User_id:        b.userId,
		Command:        notifyFORCE_BUY,
		Stock_id:       b.stock,
		Stock_price:    b.amount,
		Cash_value:     0,
	}
	return nil
}

func (b FORCE_BUY) Notify() Notification {
	return Notification{
		Topic:     notifyFORCE_BUY,
		Timestamp: time.Now(),
		Ticket:    b.ticket,
		Userid:    b.userId,
		Stock:     &b.stock,
		Amount:    &b.amount,
	}
}

func (b FORCE_BUY) Postrequsite(mb *MessageBus) error {
	return nil
}

// Purpose:
//
//	Commit Sell the specified dollar mount of the stock currently held by the
//	specified user at the current price.
//
// Pre-conditions:
//
//	The user's account must be greater or equal to the amount of the
//	purchase.
//
// Post-Conditions:
//
//	(a) the user's account for the given stock is decremented by the
//	sale amount
//	(b) the user's cash account is increased by the sell amount
//
// Example:
//
//	FORCE_SELL,jsmith,ABC,200.00
type FORCE_SELL struct {
	ticket int64
	userId string
	stock  string
	amount float64
}

func (b *FORCE_SELL) Prerequsite(rdb *redis.Client) error {
	return nil
}

func (b FORCE_SELL) Execute(ch chan *Transaction) error {
	ch <- &Transaction{
		Transaction_id: b.ticket,
		User_id:        b.userId,
		Command:        notifyFORCE_SELL,
		Stock_id:       b.stock,
		Stock_price:    b.amount,
		Cash_value:     0,
	}
	return nil
}

func (b FORCE_SELL) Notify() Notification {
	return Notification{
		Topic:     notifyFORCE_SELL,
		Timestamp: time.Now(),
		Ticket:    b.ticket,
		Userid:    b.userId,
		Stock:     &b.stock,
		Amount:    &b.amount,
	}
}

func (b FORCE_SELL) Postrequsite(mb *MessageBus) error {
	return nil
}
