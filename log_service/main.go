package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type LogType interface {
	SetTimestamp()
}

func (u *userCommand) SetTimestamp() {
	u.Timestamp = timestamp()
}

func (u *accountTransaction) SetTimestamp() {
	u.Timestamp = timestamp()
}
func (u *dumplogEvent) SetTimestamp() {
	u.Timestamp = timestamp()
}
func (u *debugEvent) SetTimestamp() {
	u.Timestamp = timestamp()
}
func (u *errorEvent) SetTimestamp() {
	u.Timestamp = timestamp()
}
func (u *quoteServer) SetTimestamp() {
	u.Timestamp = timestamp()
}
func (u *systemEvent) SetTimestamp() {
	u.Timestamp = timestamp()
}

type userCommand struct {
	//XmlName      xml.Name `xml:"userCommand"`
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Ticketnumber int      `xml:"transactionNum" json:"ticketnumber"`
	Command      []string `xml:"command" json:"command"`
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
}

type accountTransaction struct {
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Ticketnumber int      `xml:"transactionNum" json:"ticketnumber"`
	Action       []string `xml:"action" json:"action"`
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
}

type quoteServer struct {
	Timestamp       int64  `xml:"timestamp"`
	ServerName      string `xml:"server" json:"server"`
	Ticketnumber    int    `xml:"transactionNum" json:"ticketnumber"`
	Price           string `xml:"price" json:"price"`
	StockSymbol     string `xml:"stockSymbol" json:"stock_symbol"`
	Username        string `xml:"username" json:"username"`
	QuoteServerTime int64  `xml:"quoteServerTime" json:"quoteServerTime"`
}

type systemEvent struct {
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Ticketnumber int      `xml:"transactionNum" json:"ticketnumber"`
	Command      []string `xml:"command,attr" json:"command"`
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
}

type errorEvent struct {
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Ticketnumber int      `xml:"transactionNum" json:"ticketnumber"`
	Command      []string `xml:"command" json:"command"`
	Username     string   `xml:"username" json:"username"`
	DebugMessage string   `xml:"errorMessage" json:"message"`
}

type debugEvent struct {
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Ticketnumber int      `xml:"transactionNum" json:"ticketnumber"`
	Command      []string `xml:"command" json:"command"`
	DebugMessage string   `xml:"debugMessage" json:"message"`
}

type dumplogEvent struct {
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Ticketnumber int      `xml:"transactionNum" json:"ticketnumber"`
	Command      []string `xml:"command" json:"command"`
	Filename     string   `xml:"filename" json:"filename"`
}

var f *os.File

//var ec *xmlwriter.ErrCollector
// var b  *bytes.Buffer
// var ec ErrCollector

func dial(url string) (*amqp.Connection, error) {
	for {
		conn, err := amqp.Dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
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

func setupLogListeners(conn *amqp.Connection, log_topic string) (<-chan amqp.Delivery, error) {
	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	// Declare a queue
	err = ch.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	err = ch.QueueBind(
		q.Name,       // queue name
		log_topic,    // routing key
		"logs_topic", // exchange
		false,
		nil)
	failOnError(err, "Failed to declare an exchange")
	failOnError(err, "Failed to declare a queue")

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name,           // Queue name
		"logger_service", // Consumer name
		true,             // Auto-acknowledge
		false,            // Exclusive
		false,            // No-local
		false,            // No-wait
		nil,              // Arguments
	)
	return msgs, err
}

func main() {
	var err error

	_, err = os.Stat("stocklog.xml")

	if !errors.Is(err, os.ErrNotExist) {

		os.Remove("stocklog.xml")
	}

	f, err = os.OpenFile("stocklog.xml", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {

		log.Fatalf("error opening file: %v", err)

	}

	fmt.Println("Running on Port number: 8004")

	conn, err := dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	userlogs, err := setupLogListeners(conn, "userlog")
	accountlogs, err := setupLogListeners(conn, "accountlog")
	quotelogs, err := setupLogListeners(conn, "quotelog")
	systemlogs, err := setupLogListeners(conn, "systemlog")
	errorlogs, err := setupLogListeners(conn, "errorlog")
	debuglogs, err := setupLogListeners(conn, "debuglog")
	dumplogs, err := setupLogListeners(conn, "dumplog")
	go func() {
		for {
			var out []byte
			select {
			case new_log := <-userlogs:
				recive_log := &userCommand{}
				json.Unmarshal(new_log.Body, recive_log)
				recive_log.Timestamp = timestamp()
				out, _ = xml.MarshalIndent(recive_log, "", "\t")
			case new_log := <-accountlogs:
				recive_log := &accountTransaction{}
				json.Unmarshal(new_log.Body, recive_log)
				recive_log.Timestamp = timestamp()
				out, _ = xml.MarshalIndent(recive_log, "", "\t")
			case new_log := <-quotelogs:
				recive_log := &quoteServer{}
				json.Unmarshal(new_log.Body, recive_log)
				recive_log.Timestamp = timestamp()
				out, _ = xml.MarshalIndent(recive_log, "", "\t")

			case new_log := <-systemlogs:
				recive_log := &systemEvent{}
				json.Unmarshal(new_log.Body, recive_log)
				recive_log.Timestamp = timestamp()
				out, _ = xml.MarshalIndent(recive_log, "", "\t")

			case new_log := <-errorlogs:
				recive_log := &errorEvent{}
				json.Unmarshal(new_log.Body, recive_log)
				recive_log.Timestamp = timestamp()
				out, _ = xml.MarshalIndent(recive_log, "", "\t")

			case new_log := <-debuglogs:
				recive_log := &debugEvent{}
				json.Unmarshal(new_log.Body, recive_log)
				recive_log.Timestamp = timestamp()
				out, _ = xml.MarshalIndent(recive_log, "", "\t")

			case new_log := <-dumplogs:
				recive_log := &dumplogEvent{}
				json.Unmarshal(new_log.Body, recive_log)
				recive_log.Timestamp = timestamp()
				recive_log.Filename = "stocklog.xml"
				out, _ = xml.MarshalIndent(recive_log, "", "\t")
			}
			f.WriteString(string(out))
			f.WriteString("\n")

		}
	}()
	handleRequests()
	f.Close()

}

const rabbitmqHOST = "10.9.0.15"

func timestamp() int64 {
	return time.Now().Unix()
}
func handleRequests() {

	http.HandleFunc("/userlog", userlog)

	http.HandleFunc("/accountlog", accountlog)

	http.HandleFunc("/quotelog", quotelog)

	http.HandleFunc("/systemlog", systemlog)

	http.HandleFunc("/errorlog", errorlog)

	http.HandleFunc("/debuglog", debuglog)

	http.HandleFunc("/dumplog", dumplog)

	log.Fatal(http.ListenAndServe(":8004", nil))

}

func dumplog(w http.ResponseWriter, r *http.Request) {

	recive_log := &dumplogEvent{}
	err := json.NewDecoder(r.Body).Decode(recive_log)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	recive_log.Timestamp = timestamp()
	recive_log.ServerName = "stoinks_server"
	recive_log.Filename = "stocklog.xml"
	//_ = xml.NewEncoder(*xmlwriter).Encode(recive_log)
	out, _ := xml.MarshalIndent(recive_log, "", "\t")
	//fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")
}

func userlog(w http.ResponseWriter, r *http.Request) {

	recive_log := &userCommand{}
	err := json.NewDecoder(r.Body).Decode(recive_log)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	recive_log.Timestamp = timestamp()
	recive_log.ServerName = "stoinks_server"
	//_ = xml.NewEncoder(*xmlwriter).Encode(recive_log)
	out, _ := xml.MarshalIndent(recive_log, "", "\t")
	//fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")

}

func accountlog(w http.ResponseWriter, r *http.Request) {

	recive_log := &accountTransaction{}
	err := json.NewDecoder(r.Body).Decode(recive_log)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	recive_log.Timestamp = timestamp()
	recive_log.ServerName = "stoinks_server"
	//_ = xml.NewEncoder(*xmlwriter).Encode(recive_log)
	out, _ := xml.MarshalIndent(recive_log, "", "\t")
	//fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")
}
func quotelog(w http.ResponseWriter, r *http.Request) {
	recive_log := &quoteServer{}
	err := json.NewDecoder(r.Body).Decode(recive_log)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	recive_log.Timestamp = timestamp()
	recive_log.ServerName = "stoinks_server"
	//_ = xml.NewEncoder(*xmlwriter).Encode(recive_log)
	out, _ := xml.MarshalIndent(recive_log, "", "\t")
	//fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")
}
func systemlog(w http.ResponseWriter, r *http.Request) {
	recive_log := &systemEvent{}
	err := json.NewDecoder(r.Body).Decode(recive_log)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	recive_log.Timestamp = timestamp()
	recive_log.ServerName = "stoinks_server"
	//_ = xml.NewEncoder(*xmlwriter).Encode(recive_log)
	out, _ := xml.MarshalIndent(recive_log, "", "\t")
	//fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")
}
func errorlog(w http.ResponseWriter, r *http.Request) {
	recive_log := &errorEvent{}
	err := json.NewDecoder(r.Body).Decode(recive_log)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	recive_log.Timestamp = timestamp()
	recive_log.ServerName = "stoinks_server"
	//_ = xml.NewEncoder(*xmlwriter).Encode(recive_log)
	out, _ := xml.MarshalIndent(recive_log, "", "\t")
	//fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")
}

func debuglog(w http.ResponseWriter, r *http.Request) {
	recive_log := &debugEvent{}
	err := json.NewDecoder(r.Body).Decode(recive_log)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		fmt.Println("Error with request format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	recive_log.Timestamp = timestamp()
	//_ = xml.NewEncoder(*xmlwriter).Encode(recive_log)
	out, _ := xml.MarshalIndent(recive_log, "", "\t")
	fmt.Println(string(out))
	// fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")
}
