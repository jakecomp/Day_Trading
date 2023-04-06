package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type quote struct {
	Stock string  `json:"stock"`
	Price float64 `json:"price"`
}

type Command struct {
	Ticket  int      `json:"ticket"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type Message struct {
	Command string
	Data    *Command
}

type Trigger struct {
	Amount  float64 `json:"amount"`
	Command string  `json:"command"`
	Stock   string  `json:"stock"`
	Price   float64 `json:"price"`
}

type TriggerKey struct {
	Stock   string
	Command string
}

type UserTriggers struct {
	User     string
	Triggers map[TriggerKey]Trigger
}
type TriggerStore struct {
	userMap map[string]UserTriggers
	lock    sync.RWMutex
}

type ThreadSafeTriggerStore struct {
}

func (sh *TriggerStore) Get(user string) (UserTriggers, bool) {
	sh.lock.RLock()
	p, ok := sh.userMap[user]
	sh.lock.RUnlock()
	return p, ok
}
func (sh *TriggerStore) Set(user string, q UserTriggers) {
	sh.lock.Lock()
	sh.userMap[user] = q
	sh.lock.Unlock()
}

var triggerStore = &TriggerStore{
	userMap: make(map[string]UserTriggers),
}

type StockHolder struct {
	stocks map[string]quote
	lock   sync.RWMutex
}

func (sh *StockHolder) GetStock(stock string) (quote, bool) {
	sh.lock.RLock()
	p, ok := sh.stocks[stock]
	sh.lock.RUnlock()
	return p, ok
}
func (sh *StockHolder) UpdateFromJson(body []byte) {
	sh.lock.Lock()
	json.Unmarshal(body, &sh.stocks)
	sh.lock.Unlock()
}

var stocks = StockHolder{
	stocks: make(map[string]quote),
}

// var stocks map[string]quote

var rabbitChannel *amqp.Channel
var queue amqp.Queue

func set_ammount(cmd string, user string, stock string, amount float64) {
	// Make Map Key (Stock, Command)
	triggerKey := TriggerKey{}
	triggerKey.Command = cmd
	triggerKey.Stock = stock

	triggerStore.lock.Lock()
	userMap := &triggerStore.userMap
	_, ok := (*userMap)[user]
	if !ok {
		(*userMap)[user] = UserTriggers{user, make(map[TriggerKey]Trigger)}
	}

	// Check if key exists
	trigger, ok := (*userMap)[user].Triggers[triggerKey]
	if ok {
		trigger.Amount = amount

	} else {
		trigger = Trigger{}
		trigger.Stock = stock
		trigger.Amount = amount
		trigger.Price = -1.0
	}

	(*userMap)[user].Triggers[triggerKey] = trigger
	triggerStore.lock.Unlock()
}

func set_trigger(cmd string, user string, stock string, price float64) {
	// Make Map Key (Stock, Command)
	triggerKey := TriggerKey{}
	triggerKey.Command = cmd
	triggerKey.Stock = stock

	triggerStore.lock.Lock()
	defer triggerStore.lock.Unlock()
	userMap := &triggerStore.userMap
	_, ok := (*userMap)[user]
	if !ok {
		return
	}

	trigger, ok := (*userMap)[user].Triggers[triggerKey]

	if ok {
		trigger.Price = price
		(*userMap)[user].Triggers[triggerKey] = trigger
	}

}

func push_trigger(user_id string, current_price float64, current_triger Trigger, queue_name string) {

	// println("TRIGGER EXECUTING!")

	// NEED TO FIGURE OUT HOW TO CREATE COMMAND PROPERLY
	string_amount := fmt.Sprintf("%f", current_triger.Amount)
	string_price := fmt.Sprintf("%f", current_price)
	cmd := &Command{0, queue_name, []string{user_id, current_triger.Stock, string_amount, string_price}}

	// println("Trigger is %s", cmd)

	// CONVERT COMMAND TO BYTES ARRAY
	command_bytes := new(bytes.Buffer)
	json.NewEncoder(command_bytes).Encode(cmd)

	err_3 := rabbitChannel.Publish(
		"",       // name
		"worker", // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(command_bytes.Bytes()),
		})
	failOnError(err_3, "COULD NOT PUSH TO FORCE BUY / SELL QUEUE")

}

func check_triggers() {

	triggerStore.lock.Lock()

	// Iterate through each user
	userMap := triggerStore.userMap
	for user_key, all_triggers := range userMap {

		// Iterate through each trigger for this user
		for trigger_key, current_trigger := range all_triggers.Triggers {

			stonk, _ := stocks.GetStock(trigger_key.Stock)
			current_price := stonk.Price
			trigger_price := current_trigger.Price

			if trigger_key.Command == "BUY" {
				if current_price <= trigger_price {
					push_trigger(user_key, current_price, current_trigger, "FORCE_BUY")

				}
			} else if trigger_key.Command == "SELL" {
				if current_price >= trigger_price {
					push_trigger(user_key, current_price, current_trigger, "FORCE_SELL")
				}
			} else {
				// fmt.Println("UNKNOWN TRIGGER COMMAND")
			}
		}
	}

	triggerStore.lock.Unlock()

}

func delete_key(cmd string, user string, stock string) {
	triggerKey := TriggerKey{stock, cmd}
	userMap := &triggerStore.userMap
	_, ok := (*userMap)[user].Triggers[triggerKey]

	if ok {
		delete((*userMap)[user].Triggers, triggerKey)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func connectQueue(conn *amqp.Connection) (amqp.Queue, *amqp.Channel) {

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		"worker", // Queue name
		false,    // Durable
		false,    // Delete when unused
		false,    // Exclusive
		false,    // No-wait
		nil,      // Arguments
	)

	failOnError(err, "FAILED TO DECLARE FORCE BUY / SELL QUEUE")

	// Declare a queue
	q, err = ch.QueueDeclare(
		"trigger", // Queue name
		false,     // Durable
		false,     // Delete when unused
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q, ch
}

func dial(url string) (*amqp.Connection, error) {
	for {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn, err
		}
		// Rabbitmq is slow to start so we might have to wait on it
		time.Sleep(time.Second * 3)
	}

}

func main() {
	//log.SetOutput(ioutil.Discard)

	// Connect to RabbitMQ server
	time.Sleep(time.Second * 15)
	conn, err := dial("amqp://guest:guest@10.9.0.10:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	queue, rabbitChannel = connectQueue(conn)
	defer rabbitChannel.Close()

	log.Println("RUNNING ON PORT 8004...")
	go clientCode(conn)
	go triggerListener(queue)
	select {}
	//handleRequests()
}

func triggerListener(queue amqp.Queue) {
	msgs, err := rabbitChannel.Consume(
		"trigger", // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	failOnError(err, "Failed to register a consumer")

	cmd := &Command{0, "NONE", []string{"TEST"}}

	go func() {
		for msg := range msgs {
			// Call function for checking stocks for updates
			err = json.Unmarshal(msg.Body, cmd)
			if err != nil {
				panic(err)
			}

			// log.Printf("Received a trigger ticket! %s", cmd.Command)

			if cmd.Command == "SET_BUY_AMOUNT" {
				amount, _ := strconv.ParseFloat(cmd.Args[2], 64)
				set_ammount("BUY", cmd.Args[0], cmd.Args[1], amount)
			} else if cmd.Command == "SET_BUY_TRIGGER" {
				price, _ := strconv.ParseFloat(cmd.Args[2], 64)
				set_trigger("BUY", cmd.Args[0], cmd.Args[1], price)
			} else if cmd.Command == "SET_SELL_AMOUNT" {
				amount, _ := strconv.ParseFloat(cmd.Args[2], 64)
				set_ammount("SELL", cmd.Args[0], cmd.Args[1], amount)
			} else if cmd.Command == "SET_SELL_TRIGGER" {
				price, _ := strconv.ParseFloat(cmd.Args[2], 64)
				set_trigger("SELL", cmd.Args[0], cmd.Args[1], price)
			} else if cmd.Command == "CANCEL_SET_BUY" {
				delete_key("BUY", cmd.Args[0], cmd.Args[1])
			} else if cmd.Command == "CANCEL_SET_SELL" {
				delete_key("SELL", cmd.Args[0], cmd.Args[1])
			}
		}
	}()

	var forever chan struct{}
	<-forever
}

// ======= From Here On This Can Be Used For Implementing The Trigger Service =====
func setupStockListener(conn *amqp.Connection) (*amqp.Queue, *amqp.Channel, error) {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	// Setup a queue that we want to subscribe to
	err = ch.ExchangeDeclare(
		"stock_prices", // name
		"fanout",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Println("Failed to declare an exchange")
		return nil, nil, err
	}

	// Create a temperary queue that will let us subscribe without
	// removing from the main queue for everyone else
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Println("Failed to declare a queue")
		return nil, nil, err
	}

	// Bind our temperary queue to the global exchange (subscribe to stock prices)
	err = ch.QueueBind(
		q.Name,         // queue name
		"",             // routing key
		"stock_prices", // exchange
		false,
		nil,
	)
	if err != nil {
		log.Println("Failed to bind a queue")
		return nil, nil, err
	}
	failOnError(err, "Failed to bind a queue")
	return &q, ch, err
}

func clientCode(conn *amqp.Connection) {
	// Create a channel for recieving stock values
	q, ch, err := setupStockListener(conn)
	if err != nil {
		log.Println(err)
		return
	}
	defer ch.Close()
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// Create a channel for requesting stocks
	//neededStocks, err := conn.Channel()
	neededStocksQ, err := rabbitChannel.QueueDeclare(
		"stock_requests", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	go pingStocksQueuer(neededStocksQ)

	// Print all the stocks we get back
	go func() {
		for d := range msgs {
			// Call function for checking stocks for updates
			stocks.UpdateFromJson(d.Body)

			triggerStore.lock.RLocker()
			defer triggerStore.lock.RUnlock()
			//Check triggers
			check_triggers()

		}
	}()
	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

	var forever chan struct{}
	<-forever
}

func pingStocksQueuer(neededStocksQ amqp.Queue) {
	for {
		err := rabbitChannel.Publish(
			"",                 // name
			neededStocksQ.Name, // routing key
			false,              // mandatory
			false,              // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("All"),
			})
		if err != nil {
			log.Println(err)
		}

		// Wait a second
		time.Sleep(time.Second * 1)
	}
}
