package main

// TODO Add a way to lookup a users balance
// used for pre req for buy and sell
// used for post req for commit commands
// TODO Add a way to lookup a stock value
// used for selling and buying
// TODO assert timeframe for commit and cancel commands
import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type CommandType string

const (
	notifyADD         CommandType = "ADD"
	notifyBUY                     = "BUY"
	notifySELL                    = "SELL"
	notifyCOMMIT_BUY              = "COMMIT_BUY"
	notifyCOMMIT_SELL             = "COMMIT_SELL"
	notifyCANCEL_BUY              = "CANCEL_BUY"
	notifyCANCEL_SELL             = "CANCEL_SELL"
	notifyFORCE_BUY               = "FORCE_BUY"
	notifyFORCE_SELL              = "FORCE_SELL"
	notifySTOCK_PRICE             = "STOCK_PRICE"
)

type Transaction struct {
	Transaction_id int64
	User_id        UserId
	Command        CommandType
	Stock_id       string
	Stock_price    float64
	Cash_value     float64
}

type CMD interface {
	UserId() UserId
	Notify() Notification
	Prerequsite(t *TransactionStore) error
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

func UserAccountManager(mb *MessageBus, notification Notification, us *UserStore, s *StockPricer) {
	// Map storing all the currently known stock prices
	var err error
	last_ticket := -1
	switch notification.Topic {
	// case t2price := <-stockPrice:
	// 	stockPrices[*t2price.Stock] = Stock{*t2price.Stock, *t2price.Amount}
	case notifyADD:
		err = addMoney(notification, us)
		last_ticket = int(notification.Ticket)
	case notifyCOMMIT_SELL:
		p := s.lookupPrice(*notification.Stock, notification.Ticket)
		err = sellStock(p.Price, notification, us)
		last_ticket = int(notification.Ticket)
	case notifyCOMMIT_BUY:
		p := s.lookupPrice(*notification.Stock, notification.Ticket)
		// Fallback if we still don't have a stock prices
		err = buyStock(p.Price, notification, us)
		last_ticket = int(notification.Ticket)
	case notifyFORCE_SELL:
		// Fallback if we still don't have a stock price
		p := s.lookupPrice(*notification.Stock, notification.Ticket)
		err = sellStock(p.Price, notification, us)
		last_ticket = int(notification.Ticket)
	case notifyFORCE_BUY:
		p := s.lookupPrice(*notification.Stock, notification.Ticket)
		// Fallback if we still don't have a stock price
		err = buyStock(p.Price, notification, us)
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
func Run(task CMD, m *MessageBus, t *TransactionStore, us *UserStore, s *StockPricer) {
	log.Println("Executing prereq for ", reflect.TypeOf(task), task)
	err := task.Prerequsite(t)
	if err != nil {
		sendErrorLog(int64(task.Notify().Ticket), fmt.Sprint("ERROR:", err))
		return
	}
	log.Println("Executed ", reflect.TypeOf(task), task)
	log.Println("Sending notification for ", reflect.TypeOf(task), task)

	n := task.Notify()
	UserAccountManager(m, n, us, s)
	m.Publish(n.Topic, n)

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
	userId UserId
	amount float64
}

func (a ADD) UserId() UserId {
	return a.userId
}
func (a BUY) UserId() UserId {
	return a.userId
}
func (a SELL) UserId() UserId {
	return a.userId
}
func (a COMMIT_BUY) UserId() UserId {
	return a.userId
}
func (a COMMIT_SELL) UserId() UserId {
	return a.userId
}
func (a CANCEL_SELL) UserId() UserId {
	return a.userId
}
func (a CANCEL_BUY) UserId() UserId {
	return a.userId
}
func (a FORCE_BUY) UserId() UserId {
	return a.userId
}

func (a FORCE_SELL) UserId() UserId {
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
func (a ADD) Prerequsite(t *TransactionStore) error {
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
	userId UserId
	stock  string
	amount float64
}

// lookup user balance
// if invalid balance return an error describing this
func (b BUY) Prerequsite(t *TransactionStore) error {
	if b.amount < 0 {
		return errors.New(fmt.Sprint("Attempt to buy a negative amount of ", b.stock, " that amount being ", b.amount, " for user ", b.userId))
	}
	n := b.Notify()
	n.Pending(t)
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
	userId UserId
	buy    Notification
}

func (b *COMMIT_BUY) Prerequsite(t *TransactionStore) error {
	n, err := t.lastPending(b.userId, notifyBUY)
	if err != nil && err != redis.Nil {
		log.Println(b, "failed to commit purchase")
		return err
	} else if n != nil && n.Userid == b.userId && n.Ticket < b.ticket {
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
	userId UserId
	buy    Notification
}

func (b *CANCEL_BUY) Prerequsite(t *TransactionStore) error {
	n, err := t.lastPending(b.userId, notifyBUY)
	if err != nil && err != redis.Nil {
		log.Println(b, "failed to canceled purchase")
		return err
	} else if n != nil && n.Userid == b.userId && n.Ticket < b.ticket {
		log.Println(b, "canceled the purchase of ", *n)
		sendDebugLog(b.ticket, fmt.Sprintln("Cancelling purchase of ", *n, " with ", b))
		return nil
	}

	return errors.New(fmt.Sprintln("No lingering buy found for ", b))
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
	userId UserId
	stock  string
	amount float64
}

func (s SELL) Prerequsite(t *TransactionStore) error {
	if s.amount < 0 {
		return errors.New(fmt.Sprint("Attempt to sell a negative amount of ", s.stock, " that amount being ", s.amount, " for user ", s.userId))
	}
	n := s.Notify()
	n.Pending(t)
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
	userId UserId
	sell   Notification
}

func (s *COMMIT_SELL) Prerequsite(t *TransactionStore) error {
	n, err := t.lastPending(s.userId, notifySELL)
	if err != nil && err != redis.Nil {
		return err
	} else if n != nil && n.Userid == s.userId && n.Ticket < s.ticket {
		log.Println(s, "commited the sale of ", *n)
		s.sell = *n
		return nil
	}

	return errors.New(fmt.Sprintln("No lingering sells found for ", s))

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
	userId UserId
	sell   Notification
}

func (s *CANCEL_SELL) Prerequsite(t *TransactionStore) error {
	n, err := t.lastPending(s.userId, notifySELL)
	if err != nil && err != redis.Nil {
		log.Println(s, "failed to find a sale to cancel")
		return err
	} else if n != nil && n.Userid == s.userId && n.Ticket < s.ticket {
		log.Println(s, "cancelled the sale of ", *n)
		s.sell = *n
		sendDebugLog(s.ticket, fmt.Sprintln("Cancelling sale of ", *n, " with ", s))
		return nil
	}
	return errors.New(fmt.Sprintln("No lingering sells found for ", s))
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
	userId UserId
	stock  string
	amount float64
}

func (b *FORCE_BUY) Prerequsite(t *TransactionStore) error {
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
	userId UserId
	stock  string
	amount float64
}

func (b *FORCE_SELL) Prerequsite(t *TransactionStore) error {
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
