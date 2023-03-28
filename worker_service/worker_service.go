package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

const DEBUG = true

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

// Used for errors and debugging
type debugEvent struct {
	Timestamp    int64
	ServerName   string   `json:"server"`
	Ticketnumber int64    `json:"ticketnumber"`
	Command      []string `json:"command"`
	Username     string   `json:"username"`
	DebugMessage string   `json:"message"`
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
	Name  string  `json:"stock"`
	Price float64 `json:"price"`
}

type User struct {
	Balance float64
	Stocks  map[string]*StockQuantity
}

func NewUser() *User {
	return &User{
		Balance: float64(0),
		Stocks:  make(map[string]*StockQuantity, 0),
	}
}

// // Internal DB of user state
// var users map[userid]*User

// Dispatch commands based on the command string given
func dispatch(cmd Command) (CMD, error) {
	log.Println("in dispatch command is ", cmd.Command, cmd.Args)
	funcLookup := map[string]func(Command) (CMD, error){
		notifyADD: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[1], 64)
			return ADD{ticket: int64(cmd.Ticket), userId: cmd.Args[0], amount: a}, err
		},
		notifyBUY: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return BUY{ticket: int64(cmd.Ticket), userId: cmd.Args[0], stock: cmd.Args[1], amount: a}, err
		},
		notifyCOMMIT_BUY: func(cmd Command) (CMD, error) {
			return &COMMIT_BUY{ticket: int64(cmd.Ticket), userId: cmd.Args[0]}, nil
		},
		notifyCANCEL_BUY: func(cmd Command) (CMD, error) {
			return &CANCEL_BUY{ticket: int64(cmd.Ticket), userId: cmd.Args[0]}, nil
		},
		notifySELL: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return SELL{ticket: int64(cmd.Ticket), userId: cmd.Args[0], stock: cmd.Args[1], amount: a}, err
		},
		notifyCOMMIT_SELL: func(cmd Command) (CMD, error) {
			return &COMMIT_SELL{ticket: int64(cmd.Ticket), userId: cmd.Args[0]}, nil
		},
		notifyCANCEL_SELL: func(cmd Command) (CMD, error) {
			return &CANCEL_SELL{ticket: int64(cmd.Ticket), userId: cmd.Args[0]}, nil
		},
		notifyFORCE_BUY: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return &FORCE_BUY{ticket: int64(cmd.Ticket), userId: cmd.Args[0], stock: cmd.Args[1], amount: a}, err
		},
		notifyFORCE_SELL: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return &FORCE_SELL{ticket: int64(cmd.Ticket), userId: cmd.Args[0], stock: cmd.Args[1], amount: a}, err
		},
	}
	f := funcLookup[cmd.Command]
	if f == nil {
		return nil, errors.New("Undefinined command" + cmd.Command)
	}
	return funcLookup[cmd.Command](cmd)
}

// TODO avoid this blocking for to avoid unnecssecary blocking on main thread
func getNextCommand(msgs <-chan amqp.Delivery) (*Command, error) {
	for {
		// Attempt Dequeue
		resp := <-msgs
		var cmd Command
		err := json.Unmarshal(resp.Body, &cmd)
		return &cmd, err
	}

}

// Used for logging anything related to a users account
func sendAccountLog(n *Notification, bal float32) {
	a := account_log{
		Username: n.Userid,
		// Funds:        fmt.Sprint(bal),
		Funds:        fmt.Sprint(*n.Amount),
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
func sendUserLog(n Notification) {
	var u user_log
	if n.Amount == nil {
		u = user_log{
			Username:     n.Userid,
			Ticketnumber: int(n.Ticket),
			Command:      []string{n.Topic},
		}

	} else {
		u = user_log{
			Username:     n.Userid,
			Funds:        fmt.Sprint(*n.Amount),
			Ticketnumber: int(n.Ticket),
			Command:      []string{n.Topic},
		}
	}
	ulog, _ := json.Marshal(u)
	bodyReader := bytes.NewReader(ulog)
	_, err := http.Post("http://10.9.0.9:8004/userlog", "application/json", bodyReader)
	if err != nil {
		log.Fatal(err)
	}
}

func sendErrorLog(ticket int64, msg string) {
	ulog, _ := json.Marshal(debugEvent{
		ServerName:   "worker",
		Ticketnumber: ticket,
		DebugMessage: msg,
	})
	bodyReader := bytes.NewReader(ulog)
	_, err := http.Post("http://10.9.0.9:8004/errorlog", "application/json", bodyReader)
	if err != nil {
		log.Fatal(err)
	}
}

func sendDebugLog(ticket int64, msg string) {
	if DEBUG {
		ulog, _ := json.Marshal(debugEvent{
			ServerName:   "worker",
			Ticketnumber: ticket,
			DebugMessage: msg,
		})
		bodyReader := bytes.NewReader(ulog)
		_, err := http.Post("http://10.9.0.9:8004/debuglog", "application/json", bodyReader)
		if err != nil {
			log.Fatal(err)
		}

	}
}

// Logs incomming commands
func commandLogger(nch <-chan Notification) {
	for {
		sendUserLog(<-nch)
	}
}

type StockQuantity struct {
	StockName string
	Quantity  int64
}

func addMoney(newMoney Notification, db *mongo.Client, ctx *context.Context) error {
	uid := userid(newMoney.Userid)

	current_user_doc, err := read_db(string(uid), true, db, *ctx)

	if current_user_doc == nil {
		db, ctx := connect()
		current_user_doc, err = read_db(string(uid), true, db, ctx)
	}

	sendDebugLog(int64(newMoney.Ticket),
		fmt.Sprint("user doc before adding money\n",
			current_user_doc, "for notification\n",
			newMoney))

	current_user_doc.Balance += float32(*newMoney.Amount)

	newMoney.Topic = "add"
	sendAccountLog(&newMoney, current_user_doc.Balance)

	sendDebugLog(int64(newMoney.Ticket),
		fmt.Sprint("user doc after adding money\n",
			current_user_doc, "for notification\n",
			newMoney))

	update_db(current_user_doc, db, *ctx)
	return err
}
func sellStock(price float64, newMoney Notification, db *mongo.Client, ctx *context.Context) error {
	uid := userid(newMoney.Userid)

	current_user_doc, err := read_db(string(uid), false, db, *ctx)

	if current_user_doc == nil {
		db, ctx := connect()
		current_user_doc, err = read_db(string(uid), false, db, ctx)
	}

	if err != nil {
		return err
	}
	stock_owned, ok := current_user_doc.Stonks[*newMoney.Stock]
	if !ok || stock_owned <= 0 {
		return errors.New(fmt.Sprint("ERROR: less than 0 stock available", *newMoney.Stock, *newMoney.Stock, " for price ", price))
	}

	sendDebugLog(int64(newMoney.Ticket),
		fmt.Sprint("user doc before sale money\n",
			current_user_doc, "for notification\n",
			newMoney))

	current_user_doc.Balance += float32(*newMoney.Amount)
	current_user_doc.Stonks[*newMoney.Stock] -= *newMoney.Amount / price

	newMoney.Topic = "add"
	sendAccountLog(&newMoney, current_user_doc.Balance)

	sendDebugLog(int64(newMoney.Ticket), fmt.Sprint("user doc after sale money\n",
		current_user_doc, "for notification\n",
		newMoney))

	update_db(current_user_doc, db, *ctx)
	return nil
}
func buyStock(price float64, newMoney Notification, db *mongo.Client, ctx *context.Context) error {
	uid := userid(newMoney.Userid)

	current_user_doc, err := read_db(string(uid), false, db, *ctx)

	if current_user_doc == nil {
		db, ctx := connect()
		current_user_doc, err = read_db(string(uid), false, db, ctx)
	}

	if err != nil {
		return err
	}
	sendDebugLog(int64(newMoney.Ticket),
		fmt.Sprint("user doc before purchase money\n",
			current_user_doc, "for notification\n",
			newMoney, " with buy amount ", *newMoney.Amount, " of stock ", *newMoney.Stock, "\n",
			"With a value of:", price))

	current_user_doc.Balance -= float32(*newMoney.Amount)
	if current_user_doc.Balance < 0 {
		return errors.New(fmt.Sprint("Negative balance is not allowed during buy for ", *newMoney.Stock, " for price ", price))
	}

	_, ok := current_user_doc.Stonks[*newMoney.Stock]
	if !ok {
		current_user_doc.Stonks[*newMoney.Stock] = *newMoney.Amount / price
	} else {
		current_user_doc.Stonks[*newMoney.Stock] += *newMoney.Amount / price
	}

	sendDebugLog(int64(newMoney.Ticket),
		fmt.Sprint("user doc after purchase money\n",
			current_user_doc,
			"for notification\n",
			newMoney,
			" with buy amount ", *newMoney.Amount, " of stock ", *newMoney.Stock, "\n",
			"With a value of:", price))

	newMoney.Topic = "remove"
	sendAccountLog(&newMoney, current_user_doc.Balance)

	update_db(current_user_doc, db, *ctx)
	return err
}

func UserAccountManager(mb *MessageBus) {
	// users = make(map[userid]*User)
	add := mb.SubscribeAll(notifyADD)
	buy := mb.SubscribeAll(notifyCOMMIT_BUY)
	sell := mb.SubscribeAll(notifyCOMMIT_SELL)
	force_buy := mb.SubscribeAll(notifyFORCE_BUY)
	force_sell := mb.SubscribeAll(notifyFORCE_SELL)

	stockPrice := mb.SubscribeAll(notifySTOCK_PRICE)
	stockPrices := make(map[string]Stock)

	// We need a starting price before we can start these operations
	// tprice := <-stockPrice

	// price := *tprice.Amount

	db, ctx := connect()
	defer db.Disconnect(ctx)
	var err error
	for {
		last_ticket := -1

		select {
		case t2price := <-stockPrice:
			stockPrices[*t2price.Stock] = Stock{*t2price.Stock, *t2price.Amount}
			sendDebugLog(int64(t2price.Ticket),
				fmt.Sprint("Stock price found for\n",
					*t2price.Stock, "with price\n",
					*t2price.Amount))

		case newMoney := <-add:
			err = addMoney(newMoney, db, &ctx)
			last_ticket = int(newMoney.Ticket)
		case sale := <-sell:
			p, ok := stockPrices[*sale.Stock]
			// Fallback if we still don't have a stock price
			if !ok {
				p = getQuote(*sale.Stock)

				sendDebugLog(int64(sale.Ticket),
					fmt.Sprint("Had to look up stock manually for",
						*sale.Stock, "and got \n",
						p.Name, "for ", p.Price))
			}
			sellStock(p.Price, sale, db, &ctx)
			last_ticket = int(sale.Ticket)
		case sale := <-force_sell:
			p, ok := stockPrices[*sale.Stock]
			// Fallback if we still don't have a stock price
			if !ok {
				p = getQuote(*sale.Stock)
				sendDebugLog(int64(sale.Ticket),
					fmt.Sprint("Had to look up stock manually for",
						*sale.Stock, "and got \n",
						p.Name, "for ", p.Price))
			}
			sellStock(p.Price, sale, db, &ctx)
			last_ticket = int(sale.Ticket)
		case purchase := <-buy:
			p, ok := stockPrices[*purchase.Stock]
			// Fallback if we still don't have a stock price
			if !ok {
				p = getQuote(*purchase.Stock)
				sendDebugLog(int64(purchase.Ticket),
					fmt.Sprint("Had to look up stock manually for",
						*purchase.Stock, "and got \n",
						p.Name, "for ", p.Price))
			}
			buyStock(p.Price, purchase, db, &ctx)
			last_ticket = int(purchase.Ticket)
		case purchase := <-force_buy:
			p, ok := stockPrices[*purchase.Stock]
			// Fallback if we still don't have a stock price
			if !ok {
				p = getQuote(*purchase.Stock)
				sendDebugLog(int64(purchase.Ticket),
					fmt.Sprint("Had to look up stock manually for",
						*purchase.Stock, "and got \n",
						p.Name, "for ", p.Price))
			}
			buyStock(p.Price, purchase, db, &ctx)
			last_ticket = int(purchase.Ticket)
		}
		if err != nil {
			sendErrorLog(int64(last_ticket), fmt.Sprint("ERROR:", err))
		}
	}
}

func getQuote(stock string) Stock {
	var stonks Stock
	rsp, err := http.Get("http://10.9.0.6:8002/" + stock)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &stonks)
	log.Print(stonks)
	return stonks
}

func monitorStock(stockName string, mb *MessageBus) {
	for {
		S := getQuote(stockName)
		mb.Publish(notifySTOCK_PRICE, Notification{
			time.Now(),
			notifySTOCK_PRICE,
			-1,
			"",
			&S.Name,
			&S.Price,
		})
		time.Sleep(time.Millisecond * 1000)
	}
}
func stockMonitor(mb *MessageBus) {
	monitoredStocks := make(map[string]*string)
	buy := mb.SubscribeAll(notifyBUY)
	sell := mb.SubscribeAll(notifySELL)
	for {
		select {
		case stock := <-sell:
			if monitoredStocks[*stock.Stock] == nil {
				monitoredStocks[*stock.Stock] = stock.Stock
				go monitorStock(*stock.Stock, mb)
			}

		case stock := <-buy:
			if monitoredStocks[*stock.Stock] == nil {
				monitoredStocks[*stock.Stock] = stock.Stock
				go monitorStock(*stock.Stock, mb)
			}
		}
	}

}

func stockPrinter(mb *MessageBus) {
	prices := mb.SubscribeAll(notifySTOCK_PRICE)
	for {
		price := <-prices
		log.Println("Stock price of ", *price.Stock, " is ", Stock{
			Name:  *price.Stock,
			Price: *price.Amount,
		})
	}

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func setupListeners(conn *amqp.Connection) (amqp.Queue, *amqp.Channel) {
	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	// Declare a queue
	q, err := ch.QueueDeclare(
		"worker", // Queue name
		false,    // Durable
		false,    // Delete when unused
		false,    // Exclusive
		false,    // No-wait
		nil,      // Arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q, ch
}
func dial(url string) (*amqp.Connection, error) {
	for {
		conn, err := amqp.Dial("amqp://guest:guest@10.9.0.10:5672/")
		if err == nil {
			return conn, err
		}
		time.Sleep(time.Second * 3)
	}

}
func main() {
	// Determine if we should use local host
	var host string
	if len(os.Args) > 1 {
		host = "localhost"
		// Disable logging by default
		fmt.Println("WARNING! HOST SET AS LOCALHOST")
	} else {
		host = "10.9.0.7"
		fmt.Println("HOST FOR WORKER SET AS " + host)
		log.SetOutput(ioutil.Discard)
	}
	// Connect to RabbitMQ server
	conn, err := dial("amqp://guest:guest@10.9.0.10:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	q, ch := setupListeners(conn)
	defer ch.Close()

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name,           // Queue name
		"worker_service", // Consumer name
		true,             // Auto-acknowledge
		false,            // Exclusive
		false,            // No-local
		false,            // No-wait
		nil,              // Arguments
	)
	failOnError(err, "Failed to register a consumer")

	// Message bus shared between commands
	mb := NewMessageBus()

	// intended for updating the DB when used
	// ch := make(chan *Transaction)
	// Used for logging commands when recieved
	// nch := make(chan *Notification)

	// Logs all transactions to user accounts
	go UserAccountManager(mb)
	// Monitor the current stock value
	go stockMonitor(mb)
	// go stockPrinter(mb)
	// Log the new command
	notes := []string{
		notifyADD,
		notifyBUY,
		notifySELL,
		notifyCOMMIT_BUY,
		notifyCOMMIT_SELL,
		notifyCANCEL_BUY,
		notifyCANCEL_SELL,
	}
	nch := make(chan Notification)
	go commandLogger(nch)

	waitChan := make(chan struct{}, MAX_CONCURRENT_JOBS)
	for _, n := range notes {
		val := n
		c := mb.SubscribeAll(val)
		waitChan <- struct{}{}
		go func() {
			// Logs all incoming commands
			<-waitChan
			for {
				r := <-c
				nch <- r
			}
		}()
	}

	for {
		select {
		default:
			t, err := getNextCommand(msgs)
			cmd, err := dispatch(*t)

			if err == nil {
				// Execute this new command
				waitChan <- struct{}{}
				go Run(cmd, mb)
				<-waitChan
				// Sleep is here to avoid blocking the
				// queue server for too long
				// time.Sleep(time.Millisecond * 1)
			} else {
				sendErrorLog(int64(t.Ticket), fmt.Sprint("ERROR:", err))
			}
		}

	}
}
