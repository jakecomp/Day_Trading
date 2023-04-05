package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	DEBUG = false

	MAX_CONCURRENT_JOBS = 300
	redisHOST           = "10.9.0.7"
	rabbitmqHOST        = "10.9.0.10"
	quoteHOST           = "10.9.0.6"
	logHOST             = "10.9.0.9"
)

type UserId string
type Args []string

type Command struct {
	Ticket  int
	Command CommandType
	Args    Args
}

type Stock struct {
	Name  string  `json:"stock"`
	Price float64 `json:"price"`
}

// Dispatch commands based on the command string given
func dispatch(cmd Command) (CMD, error) {
	log.Println("in dispatch command is ", cmd.Command, cmd.Args)
	funcLookup := map[CommandType]func(Command) (CMD, error){
		notifyADD: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[1], 64)
			return ADD{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0]), amount: a}, err
		},
		notifyBUY: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return BUY{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0]), stock: cmd.Args[1], amount: a}, err
		},
		notifyCOMMIT_BUY: func(cmd Command) (CMD, error) {
			return &COMMIT_BUY{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0])}, nil
		},
		notifyCANCEL_BUY: func(cmd Command) (CMD, error) {
			return &CANCEL_BUY{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0])}, nil
		},
		notifySELL: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return SELL{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0]), stock: cmd.Args[1], amount: a}, err
		},
		notifyCOMMIT_SELL: func(cmd Command) (CMD, error) {
			return &COMMIT_SELL{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0])}, nil
		},
		notifyCANCEL_SELL: func(cmd Command) (CMD, error) {
			return &CANCEL_SELL{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0])}, nil
		},
		notifyFORCE_BUY: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return &FORCE_BUY{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0]), stock: cmd.Args[1], amount: a}, err
		},
		notifyFORCE_SELL: func(cmd Command) (CMD, error) {
			a, err := strconv.ParseFloat(cmd.Args[2], 64)
			return &FORCE_SELL{ticket: int64(cmd.Ticket), userId: UserId(cmd.Args[0]), stock: cmd.Args[1], amount: a}, err
		},
	}
	f := funcLookup[cmd.Command]
	if f == nil {
		return nil, errors.New("Undefinined command" + string(cmd.Command))
	}
	return funcLookup[cmd.Command](cmd)
}

func getNextCommand(resp amqp.Delivery) (*[]Command, error) {
	// Attempt Dequeue
	var cmd []Command
	err := json.Unmarshal(resp.Body, &cmd)
	return &cmd, err
}

// we return a function so that we can block during the subscribing process
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
		conn, err := amqp.Dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
		if err == nil {
			return conn, err
		}
		time.Sleep(time.Second * 3)
	}
}

func main() {

	// Disable printing logs
	log.SetOutput(ioutil.Discard)

	// Connect to RabbitMQ server
	conn, err := dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
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

	// Monitor the current stock value
	// go stockMonitor(mb)
	// Log the new command

	// Setup For Logging User Commands
	startCommandLogger(mb)

	logchannel = setupLogger()

	// Used for tracking the last seen price of a stock
	stock_pricer := &StockPricer{ // Maybe replace with redis?
		rdb: setupRedis(),
		// prices: make(map[string]Stock),
	}
	// Stores pending Buy and Sells in redis for us
	pendingTransactions := NewTransactionStore()
	defer pendingTransactions.ShutDown()

	users := NewUserStore()
	defer users.Disconnect()

	limitThreadCount := make(chan struct{}, MAX_CONCURRENT_JOBS)

	for {
		select {
		case next := <-msgs:
			nextUser, err := getNextCommand(next)
			if err == nil {
				limitThreadCount <- struct{}{}
				go func(cmds []Command) {
					for _, t := range cmds {
						cmd, err := dispatch(t)
						if err == nil {
							// Execute this new command
							Run(cmd, mb, pendingTransactions, users, stock_pricer)
						} else {
							sendErrorLog(int64(t.Ticket), fmt.Sprint("ERROR:", err))
						}
					}
					<-limitThreadCount
				}(*nextUser)

			}
		case <-time.After(time.Second * 5):
			fmt.Println("5 seconds since last message")
		}

	}
}
