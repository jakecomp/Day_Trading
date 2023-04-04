package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

type quote struct {
	Stock string  `json:"stock"`
	Price float64 `json:"price"`
}

type command struct {
	Ticket  int
	Command string
	Args    []string
}

type Message struct {
	Command string
	Data    *command
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

var userMap map[string]UserTriggers
var stocks map[string]quote

var upgrader = websocket.Upgrader{}
var rabbitChannel *amqp.Channel
var queue amqp.Queue

// func handleRequests() {
// 	//http.HandleFunc("/ws", socketHandler)
// 	log.Fatal(http.ListenAndServe(":8004", nil))
// }

// func socketHandler(w http.ResponseWriter, r *http.Request) {
// 	// Upgrade our raw HTTP connection to a websocket based one
// 	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Print("Error during connection upgradation:", err)
// 		return
// 	}
// 	defer conn.Close()
// 	socketReader(conn)
// }

func set_ammount(cmd string, user string, stock string, amount float64) {
	// Make Map Key (Stock, Command)
	triggerKey := TriggerKey{}
	triggerKey.Command = cmd
	triggerKey.Stock = stock

	// Check if key exists
	trigger, ok := userMap[user].Triggers[triggerKey]
	if ok {
		trigger.Amount = amount

	} else {
		trigger = Trigger{}
		trigger.Stock = stock
		trigger.Amount = amount
		trigger.Price = -1.0
	}

	userMap[user].Triggers[triggerKey] = trigger
}

func set_trigger(cmd string, user string, stock string, price float64) {
	// Make Map Key (Stock, Command)
	triggerKey := TriggerKey{}
	triggerKey.Command = cmd

	triggerKey.Stock = stock

	trigger, ok := userMap[user].Triggers[triggerKey]

	if ok {
		trigger.Price = price
		userMap[user].Triggers[triggerKey] = trigger
	}
}

func push_trigger(user_id string, current_price float64, current_triger Trigger, queue_name string) {

	conn, _ := dial("amqp://guest:guest@10.9.0.10:5672/")
	ch, err := conn.Channel()
	failOnError(err, "FAILED TO CONNECT TO rabbitMQ")

	q, err_2 := ch.QueueDeclare(
		queue_name, // Queue name
		false,      // Durable
		false,      // Delete when unused
		false,      // Exclusive
		false,      // No-wait
		nil,        // Arguments
	)

	failOnError(err_2, "FAILED TO DECLARE FORCE BUY / SELL QUEUE")

	// NEED TO FIGURE OUT HOW TO CREATE COMMAND PROPERLY
	string_amount := fmt.Sprintf("%f", current_triger.Amount)
	string_price := fmt.Sprintf("%f", current_price)
	cmd := &command{0, queue_name, []string{user_id, string_amount, string_price, current_triger.Stock}}

	// CONVERT COMMAND TO BYTES ARRAY
	command_bytes := new(bytes.Buffer)
	json.NewEncoder(command_bytes).Encode(cmd)

	err_3 := ch.Publish(
		"",     // name
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(command_bytes.Bytes()),
		})
	failOnError(err_3, "COULD NOT PUSH TO FORCE BUY / SELL QUEUE")

}

func check_triggers() {

	// Iterate through each user

	for user_key, all_triggers := range userMap {

		// Iterate through each trigger for this user
		for trigger_key, current_trigger := range all_triggers.Triggers {

			current_price := stocks[trigger_key.Stock].Price
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
				fmt.Println("UNKNOWN TRIGGER COMMAND")
			}
		}
	}
}

func delete_key(cmd string, user string, stock string) {
	triggerKey := TriggerKey{stock, cmd}
	delete(userMap[user].Triggers, triggerKey)
}

// Grab quotes every second and compare to quote values
// func socketReader(conn *websocket.Conn) {
// 	// Event Loop, Handle Comms in here
// 	cmd := &command{0, "NONE", []string{"TEST"}}
// 	for {
// 		// Recieve a trigger command
// 		_, message, err := conn.ReadMessage()
// 		err = json.Unmarshal(message, cmd)
// 		if err != nil {
// 			fmt.Println("Error during message reading:", err)
// 			break
// 		}

// 		if cmd.Command == "SET_BUY_AMOUNT" {
// 			amount, _ := strconv.ParseFloat(cmd.Args[2], 64)
// 			set_ammount("BUY", cmd.Args[0], cmd.Args[1], amount)
// 		} else if cmd.Command == "SET_BUY_TRIGGER" {
// 			price, _ := strconv.ParseFloat(cmd.Args[2], 64)
// 			set_trigger("BUY", cmd.Args[0], cmd.Args[1], price)
// 		} else if cmd.Command == "SET_SELL_AMOUNT" {
// 			amount, _ := strconv.ParseFloat(cmd.Args[2], 64)
// 			set_ammount("SELL", cmd.Args[0], cmd.Args[1], amount)
// 		} else if cmd.Command == "SET_SELL_TRIGGER" {
// 			price, _ := strconv.ParseFloat(cmd.Args[2], 64)
// 			set_trigger("SELL", cmd.Args[0], cmd.Args[1], price)
// 		} else if cmd.Command == "CANCEL_SET_BUY" {
// 			delete_key("BUY", cmd.Args[0], cmd.Args[1])
// 		} else if cmd.Command == "CANCEL_SET_SELL" {
// 			delete_key("SELL", cmd.Args[0], cmd.Args[1])
// 		}
// 	}
// }

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func connectQueue(conn *amqp.Connection) (amqp.Queue, *amqp.Channel) {

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	// Declare a queue
	q, err := ch.QueueDeclare(
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

	userMap = make(map[string]UserTriggers)

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

	go func() {
		for msg := range msgs {
			// Call function for checking stocks for updates

			log.Printf("Received a trigger ticket! %s", msg.Body)
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

			json.Unmarshal(d.Body, &stocks)

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
