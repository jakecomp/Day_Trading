package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"time"

	"github.com/streadway/amqp"
)

const (
	mqhost    = "10.9.0.10"
	quotehost = "10.9.0.6"
)

// If you wanna run this locally
// const (
// 	mqhost    = "localhost"
// 	quotehost = "localhost"
// )

type Stock struct {
	Name  string  `json:"stock"`
	Price float64 `json:"price"`
}

var stock_map map[string]Stock

func dial(url string) (*amqp.Connection, error) {
	for {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn, err
		}
		time.Sleep(time.Second * 3)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func getQuote(stock string) Stock {
	var stonks Stock
	rsp, err := http.Get("http://" + quotehost + ":8002")
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

func getAllQuotes() map[string]Stock {
	var stonks map[string]Stock
	rsp, err := http.Get("http://" + quotehost + ":8002/all")
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

func stockMonitor(stockReqs <-chan string, stockResults chan Stock) {
	for stock := range stockReqs {
		stockName := stock
		go func() { stockResults <- getQuote(stockName) }()
	}
}

func main() {
	conn, err := dial("amqp://guest:guest@" + mqhost + ":5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"stock_prices", // name
		"fanout",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to create exchange in RabbitMQ")

	// input from clients
	inputCh, err := conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ")
	defer inputCh.Close()

	failOnError(err, "Failed to create exchange in RabbitMQ")
	inputQ, err := inputCh.QueueDeclare(
		"stock_requests", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	requests, err := inputCh.Consume(
		inputQ.Name, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // arg
	)
	//go clientCode(conn)
	// Start the goroutine to request stocks
	for request := range requests {
		var s interface{}
		if string(request.Body) == "All" {
			s = getAllQuotes()
		} else {
			s = getQuote(string(request.Body))
		}
		body, err := json.Marshal(s)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}
		err = ch.Publish(
			"stock_prices", // exchange
			"",             // routing key
			false,          // mandatory
			false,          // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        body,
			})
	}
}

// ======= From Here On This Can Be Used For Implementing The Trigger Service =====
// func setupStockListener(conn *amqp.Connection) (*amqp.Queue, *amqp.Channel, error) {
// 	ch, err := conn.Channel()
// 	failOnError(err, "Failed to open a channel")

// 	// Setup a queue that we want to subscribe to
// 	err = ch.ExchangeDeclare(
// 		"stock_prices", // name
// 		"fanout",       // type
// 		true,           // durable
// 		false,          // auto-deleted
// 		false,          // internal
// 		false,          // no-wait
// 		nil,            // arguments
// 	)
// 	if err != nil {
// 		log.Println("Failed to declare an exchange")
// 		return nil, nil, err
// 	}

// 	// Create a temperary queue that will let us subscribe without
// 	// removing from the main queue for everyone else
// 	q, err := ch.QueueDeclare(
// 		"",    // name
// 		false, // durable
// 		false, // delete when unused
// 		true,  // exclusive
// 		false, // no-wait
// 		nil,   // arguments
// 	)
// 	if err != nil {
// 		log.Println("Failed to declare a queue")
// 		return nil, nil, err
// 	}

// 	// Bind our temperary queue to the global exchange (subscribe to stock prices)
// 	err = ch.QueueBind(
// 		q.Name,         // queue name
// 		"",             // routing key
// 		"stock_prices", // exchange
// 		false,
// 		nil,
// 	)
// 	if err != nil {
// 		log.Println("Failed to bind a queue")
// 		return nil, nil, err
// 	}
// 	failOnError(err, "Failed to bind a queue")
// 	return &q, ch, err
// }
// func clientCode(conn *amqp.Connection) {
// 	// Create a channel for recieving stock values
// 	q, ch, err := setupStockListener(conn)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	defer ch.Close()
// 	msgs, err := ch.Consume(
// 		q.Name, // queue
// 		"",     // consumer
// 		true,   // auto-ack
// 		false,  // exclusive
// 		false,  // no-local
// 		false,  // no-wait
// 		nil,    // args
// 	)
// 	failOnError(err, "Failed to register a consumer")

// 	// Create a channel for requesting stocks
// 	neededStocks, err := conn.Channel()
// 	neededStocksQ, err := neededStocks.QueueDeclare(
// 		"stock_requests", // name
// 		false,            // durable
// 		false,            // delete when unused
// 		false,            // exclusive
// 		false,            // no-wait
// 		nil,              // arguments
// 	)
// 	// Example of requesting 10 instances of the stock "S"
// 	for i := 0; i < 10; i++ {

// 		err = neededStocks.Publish(
// 			"",                 // name
// 			neededStocksQ.Name, // routing key
// 			false,              // mandatory
// 			false,              // immediate
// 			amqp.Publishing{
// 				ContentType: "text/plain",
// 				Body:        []byte("S"),
// 			})
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	}
// 	// Print all the stocks we get back
// 	go func() {
// 		for d := range msgs {
// 			log.Printf(" [x] got %s", d.Body)
// 		}
// 	}()
// 	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

// 	var forever chan struct{}
// 	<-forever
// }
