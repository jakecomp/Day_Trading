package main

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
)

type Transaction struct {
	Transaction_id string
	User_id        string
	Command        string
	Stock_id       string
	Stock_price    float64
	Cash_value     float64
}

type CMD interface {
	Notify(*MessageBus)
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
	go c.Notify(m)
	log.Println("Notification sent for ", reflect.TypeOf(c), c)
	log.Println("Waiting on Postreq for ", reflect.TypeOf(c), c)
	log.Println("Executing Postreq for ", reflect.TypeOf(c), c)
	err = c.Postrequsite(m)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Executed Postreq for ", reflect.TypeOf(c), c)
}

const (
	notifyGET_BALANCE = "GET_BALANCE"
	notifyBALANCE     = "BALANCE"
	notifyADD         = "ADD"
	notifyBUY         = "BUY"
	notifySELL        = "SELL"
	notifyCOMMIT_BUY  = "COMMIT_BUY"
	notifyCOMMIT_SELL = "COMMIT_SELL"
	notifyCANCEL_BUY  = "CANCEL_BUY"
	notifyCANCEL_SELL = "CANCEL_SELL"
)

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
	userId string
	amount float64
}

func (a ADD) Notify(mb *MessageBus) {
	mb.Publish(notifyADD, Notification{
		Timestamp: time.Now(),
		Userid:    a.userId,
		Stock:     nil,
		Amount:    &a.amount,
	})
}

// lookup user balance
// if invalid balance return an error describing this
func (a ADD) Prerequsite(mb *MessageBus) error {
	return nil
}
func (a ADD) Execute(ch chan *Transaction) error {
	ch <- &Transaction{Transaction_id: "ID_1", User_id: a.userId, Command: notifyADD, Cash_value: a.amount, Stock_id: "", Stock_price: 0}
	return nil
}

// TODO
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
	userId string
	stock  string
	amount float64
	cost   float64
}

// lookup user balance
// if invalid balance return an error describing this
func (b BUY) Prerequsite(mb *MessageBus) error {
	// TODO determine how to lookup user balance
	// TODO get the current price of that stock
	ch := mb.Subscribe(notifyADD, userid(b.userId))
	// TODO or balance
	for n := range ch {
		if n.Userid == b.userId {
			if *n.Amount < b.cost {
				return errors.New("Not enough money for this stock")
			} else {
				return nil
			}
		}
	}
	return errors.New("Balance Channel Prematurely Closed")

}

// Removed since nothing should actually happen until the commit or cancel
func (b BUY) Execute(ch chan *Transaction) error {
	return nil
}

func (b BUY) Notify(mb *MessageBus) {
	mb.Publish(notifyBUY, Notification{
		Timestamp: time.Now(),
		Userid:    b.userId,
		Stock:     &b.stock,
		Amount:    &b.cost,
	})
}

func (b BUY) Postrequsite(mb *MessageBus) error {
	// TODO determine how to lookup user balance
	commitChan := mb.Subscribe(notifyCOMMIT_BUY, userid(b.userId))
	cancelChan := mb.Subscribe(notifyCANCEL_BUY, userid(b.userId))

	select {
	case n := <-commitChan:
		if n.Userid == b.userId {
			return nil
		}

	case n := <-cancelChan:
		if n.Userid == b.userId {
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
	userId string
	buy    Notification
}

// TODO Only take from the backlog
// TODO assert that it is within the last 60 seconds
//
//	and not later than when the commit was requested
func (b *COMMIT_BUY) Prerequsite(mb *MessageBus) error {
	ch := mb.Subscribe(notifyBUY, userid(b.userId))

	for n := range ch {
		if n.Userid == b.userId {
			b.buy = n
			return nil
		}
	}
	return errors.New("Balance Channel Prematurely Closed")

}

// TODO
func (b COMMIT_BUY) Execute(ch chan *Transaction) error {
	// TODO get these values
	ch <- &Transaction{
		Transaction_id: "ID_1",
		User_id:        b.userId,
		Command:        notifyCOMMIT_BUY,
		Stock_id:       *b.buy.Stock,
		Stock_price:    24.5,
		Cash_value:     600.0,
	}
	return nil
}

func (b COMMIT_BUY) Notify(mb *MessageBus) {
	mb.Publish(notifyCOMMIT_BUY, Notification{
		Timestamp: time.Now(),
		Userid:    b.userId,
		Stock:     b.buy.Stock,
		Amount:    b.buy.Amount,
	})
}

// TODO
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
	userId string
	buy    Notification
}

// TODO Only take from the backlog
func (b *CANCEL_BUY) Prerequsite(mb *MessageBus) error {
	ch := mb.Subscribe(notifyBUY, userid(b.userId))

	for n := range ch {
		if n.Userid == b.userId {
			b.buy = n
			return nil
		}
	}
	return errors.New("Balance Channel Prematurely Closed")

}

// TODO
func (b CANCEL_BUY) Execute(ch chan *Transaction) error {
	return nil
}

func (b CANCEL_BUY) Notify(mb *MessageBus) {
	mb.Publish(notifyCANCEL_BUY, Notification{
		Timestamp: time.Now(),
		Userid:    b.userId,
		Stock:     b.buy.Stock,
		Amount:    b.buy.Amount,
	})
}

// TODO
func (b CANCEL_BUY) Postrequsite(mb *MessageBus) error {
	// TODO determine how to lookup user balance
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
	userId string
	stock  string
	amount float64
	cost   float64
}

// TODO check amount of stock held by user
// TODO
func (s SELL) Prerequsite(mb *MessageBus) error {
	// Wait for the user to buy this stock
	ch := mb.Subscribe(notifyCOMMIT_BUY, userid(s.userId))

	for n := range ch {
		if n.Userid == s.userId {
			if *n.Amount < s.amount {
				return errors.New(fmt.Sprintf("Don't have %f of %s", s.amount, s.stock))
			} else {
				return nil
			}
		}
	}
	return errors.New("Balance Channel Prematurely Closed")
}
func (b SELL) Execute(ch chan *Transaction) error {
	return nil
}

func (s SELL) Notify(mb *MessageBus) {
	mb.Publish(notifySELL, Notification{
		Timestamp: time.Now(),
		Userid:    s.userId,
		Stock:     &s.stock,
		Amount:    &s.cost,
	})
}

func (s SELL) Postrequsite(mb *MessageBus) error {
	// TODO determine how to lookup user balance
	commitChan := mb.Subscribe(notifyCOMMIT_SELL, userid(s.userId))
	cancelChan := mb.Subscribe(notifyCANCEL_SELL, userid(s.userId))
	select {
	case n := <-commitChan:
		if n.Userid == s.userId {
			return nil
		}

	case n := <-cancelChan:
		if n.Userid == s.userId {
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
	userId string
	sell   Notification
}

// TODO assert timeframe
func (s *COMMIT_SELL) Prerequsite(mb *MessageBus) error {
	ch := mb.Subscribe(notifySELL, userid(s.userId))

	for n := range ch {
		if n.Userid == s.userId {
			s.sell = n
			return nil
		}
	}
	return errors.New("Balance Channel Prematurely Closed")

}

// TODO
func (s COMMIT_SELL) Execute(ch chan *Transaction) error {
	// TODO get these values
	ch <- &Transaction{
		Transaction_id: "ID_1",
		User_id:        s.userId,
		Command:        notifyCOMMIT_SELL,
		Stock_id:       *s.sell.Stock,
		Stock_price:    24.5,
		Cash_value:     600.0,
	}
	return nil
}

func (s COMMIT_SELL) Notify(mb *MessageBus) {
	mb.Publish(notifyCOMMIT_SELL, Notification{
		Timestamp: time.Now(),
		Userid:    s.userId,
		Stock:     s.sell.Stock,
		Amount:    s.sell.Amount,
	})
}

// TODO
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
//
// TODO implement cancelation
type CANCEL_SELL struct {
	userId string
	sell   Notification
}

// TODO assert timeframe
func (s *CANCEL_SELL) Prerequsite(mb *MessageBus) error {
	ch := mb.Subscribe(notifySELL, userid(s.userId))

	for n := range ch {
		if n.Userid == s.userId {
			s.sell = n
			return nil
		}
	}
	return errors.New("Balance Channel Prematurely Closed")

}

// TODO
func (s CANCEL_SELL) Execute(ch chan *Transaction) error {
	return nil
}

func (s CANCEL_SELL) Notify(mb *MessageBus) {
	mb.Publish(notifyCANCEL_SELL, Notification{
		Timestamp: time.Now(),
		Userid:    s.userId,
		Stock:     s.sell.Stock,
		Amount:    s.sell.Amount,
	})
}

// TODO
func (b CANCEL_SELL) Postrequsite(mb *MessageBus) error {
	return nil
}

func main() {
	{
		ch := make(chan *Transaction)
		go func() {
			for c := range ch {
				log.Println("Transaction commited for", c.Command)
			}
		}()

		mb := NewMessageBus()
		addcmd := ADD{userId: "me", amount: 32.1}
		buycmd := BUY{userId: "me", stock: "ABC", cost: 32.1, amount: 1.0}
		buycmd_cmt := &COMMIT_BUY{userId: "me"}
		go Run(addcmd, mb, ch)
		go Run(buycmd, mb, ch)
		go Run(buycmd_cmt, mb, ch)

	}
}
