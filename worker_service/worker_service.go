package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	redis "github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func pendingPurchaseCacher(mb *MessageBus, rdb *redis.Client) {
	// TODO cancel purchases
}

const (
	redisHOST    = "10.9.0.7"
	rabbitmqHOST = "10.9.0.10"
	quoteHOST    = "10.9.0.6"
	logHOST      = "10.9.0.9"
)

func setupRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		DB:       0,
		Password: "",
		Addr:     redisHOST + ":6379",
	})
	return client
}

func (i *Notification) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

func (b *Notification) Pending(client *redis.Client) error {
	ctx := context.Background()
	err := client.Set(ctx, b.Userid+"#"+b.Topic, b, 0).Err()
	return err
}

func lastPending(userid string, topic string, client *redis.Client) (*Notification, error) {

	ctx := context.Background()
	val, err := client.GetDel(ctx, userid+"#"+topic).Bytes()
	if err != nil {
		return nil, err
	}

	var n Notification
	err = json.Unmarshal(val, &n)
	return &n, err

}

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

// Internal DB of user state
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
func getNextCommand(msgs <-chan amqp.Delivery) (*[]Command, error) {
	// Attempt Dequeue
	resp := <-msgs
	var cmd []Command
	err := json.Unmarshal(resp.Body, &cmd)
	log.Println("we got", cmd)
	return &cmd, err
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
	_, err := http.Post("http://"+logHOST+":8004/accountlog", "application/json", bodyReader)
	if err != nil {
		log.Println(err)
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
	_, err := http.Post("http://"+logHOST+":8004/userlog", "application/json", bodyReader)
	if err != nil {
		log.Println(err)
	}
}

func sendErrorLog(ticket int64, msg string) {
	ulog, _ := json.Marshal(debugEvent{
		ServerName:   "worker",
		Ticketnumber: ticket,
		DebugMessage: msg,
	})
	bodyReader := bytes.NewReader(ulog)
	_, err := http.Post("http://"+logHOST+":8004/errorlog", "application/json", bodyReader)
	if err != nil {
		log.Println(err)
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
		log.Println(string(ulog))
		_, err := http.Post("http://"+logHOST+":8004/debuglog", "application/json", bodyReader)
		if err != nil {
			log.Println(err)
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

func lookupUser(uid userid, db *mongo.Client, ctx *context.Context) (*user_doc, error) {
	return read_db(string(uid), false, db, *ctx)
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
func lookupPrice(stock string, ticket int64, stockPrices *map[string]Stock) Stock {
	p, ok := (*stockPrices)[stock]
	// Fallback if we still don't have a stock price
	if !ok {
		p = getQuote(stock)
		(*stockPrices)[stock] = p
		sendDebugLog(int64(ticket),
			fmt.Sprint("Had to look up stock manually for",
				stock, "and got \n",
				p.Name, "for ", p.Price))
	}
	return p

}

// we return a function so that we can block during the subscribing process
func UserAccountManager(mb *MessageBus, uid userid) func(chan struct{}) {
	add := mb.Subscribe(notifyADD, uid)
	buy := mb.Subscribe(notifyCOMMIT_BUY, uid)
	sell := mb.Subscribe(notifyCOMMIT_SELL, uid)
	force_buy := mb.Subscribe(notifyFORCE_BUY, uid)
	force_sell := mb.Subscribe(notifyFORCE_SELL, uid)
	stockPrice := mb.Subscribe(notifySTOCK_PRICE, uid)
	// Map storing all the currently known stock prices
	return func(waitChan chan struct{}) {
		waitChan <- struct{}{}
		stockPrices := make(map[string]Stock)
		db, ctx := connect()
		defer db.Disconnect(ctx)
		var err error
		for {
			last_ticket := -1
			select {
			case t2price := <-stockPrice:
				stockPrices[*t2price.Stock] = Stock{*t2price.Stock, *t2price.Amount}
			case newMoney := <-add:
				err = addMoney(newMoney, db, &ctx)
				last_ticket = int(newMoney.Ticket)
			case sale := <-sell:
				p := lookupPrice(*sale.Stock, sale.Ticket, &stockPrices)
				err = sellStock(p.Price, sale, db, &ctx)
				last_ticket = int(sale.Ticket)
			case purchase := <-buy:
				p := lookupPrice(*purchase.Stock, purchase.Ticket, &stockPrices)
				// Fallback if we still don't have a stock prices
				err = buyStock(p.Price, purchase, db, &ctx)
				last_ticket = int(purchase.Ticket)
			case sale := <-force_sell:
				// Fallback if we still don't have a stock price
				p := lookupPrice(*sale.Stock, sale.Ticket, &stockPrices)
				err = sellStock(p.Price, sale, db, &ctx)
				last_ticket = int(sale.Ticket)
			case purchase := <-force_buy:
				p := lookupPrice(*purchase.Stock, purchase.Ticket, &stockPrices)
				// Fallback if we still don't have a stock price
				err = buyStock(p.Price, purchase, db, &ctx)
				last_ticket = int(purchase.Ticket)
			default:

			}
			if err != nil {
				sendErrorLog(int64(last_ticket), fmt.Sprint("ERROR:", err))
			}
			<-waitChan
		}
	}
}

func getQuote(stock string) Stock {
	var stonks Stock
	rsp, err := http.Get("http://" + quoteHOST + ":8002/" + stock)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &stonks)
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
			s := stock
			if monitoredStocks[*s.Stock] == nil {
				monitoredStocks[*s.Stock] = s.Stock
				go monitorStock(*s.Stock, mb)
			}

		case stock := <-buy:
			s := stock
			if monitoredStocks[*s.Stock] == nil {
				monitoredStocks[*s.Stock] = s.Stock
				go monitorStock(*s.Stock, mb)
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
		// log.SetOutput(ioutil.Discard)
	}
	// Connect to RabbitMQ server
	conn, err := dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	q, ch := setupListeners(conn)
	defer ch.Close()
	rdb := setupRedis()
	defer rdb.Close()

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

	// Monitor the current stock value
	// go stockMonitor(mb)
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

	for _, n := range notes {
		val := n
		c := mb.SubscribeAll(val)
		go func() {
			// Logs all incoming commands
			for {
				r := <-c
				nch <- r
			}
		}()
	}

	waitChan := make(chan struct{}, MAX_CONCURRENT_JOBS)
	// taskChan := make(chan CMD, MAX_CONCURRENT_JOBS)

	db, ctx := connect()
	defer db.Disconnect(ctx)
	stock_pricer := &StockPricer{
		prices: make(map[string]Stock),
	}
	for {
		nextUser, err := getNextCommand(msgs)
		if err == nil {
			waitChan <- struct{}{}
			go func(cmds []Command) {
				for _, t := range cmds {
					cmd, err := dispatch(t)
					if err == nil {
						// Execute this new command
						Run(cmd, mb, rdb, db, ctx, stock_pricer)
					} else {
						sendErrorLog(int64(t.Ticket), fmt.Sprint("ERROR:", err))
					}
				}
				<-waitChan
			}(*nextUser)

		}

	}
}
