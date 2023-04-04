package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	// "net/http"
)

var logchannel *amqp.Channel

func sendLog(topic string, content []byte) error {
	return logchannel.Publish(
		"logs_topic", // exchange
		topic,        // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        content,
		})
}
func setupLogger() *amqp.Channel {
	conn, err := dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	ch, err := conn.Channel()
	failOnError(err, "Failed to declare an exchange")
	err = ch.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to publish a message")
	logchannel = ch
	return ch
}

// Logs User Commands
func startCommandLogger(mb *MessageBus) {
	notes := []CommandType{
		notifyADD,
		notifyBUY,
		notifySELL,
		notifyCOMMIT_BUY,
		notifyCOMMIT_SELL,
		notifyCANCEL_BUY,
		notifyCANCEL_SELL,
	}

	nch := make(chan Notification)

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
	go func() {
		for {
			sendUserLog(<-nch)
		}
	}()
}

// User command logs
func sendUserLog(n Notification) {
	var u user_log
	if n.Amount == nil {
		u = user_log{
			Username:     n.Userid,
			Ticketnumber: int(n.Ticket),
			Command:      []string{string(n.Topic)},
		}

	} else {
		u = user_log{
			Username:     n.Userid,
			Funds:        fmt.Sprint(*n.Amount),
			Ticketnumber: int(n.Ticket),
			Command:      []string{string(n.Topic)},
		}
	}
	ulog, _ := json.Marshal(u)
	// bodyReader := bytes.NewReader(ulog)
	// _, err := http.Post("http://"+logHOST+":8004/userlog", "application/json", bodyReader)
	err := sendLog("userlog", ulog)
	if err != nil {
		log.Println(err)
	}
}

// Used for logging anything related to a users account
func sendAccountLog(n *Notification, bal float32) {
	a := account_log{
		Username: n.Userid,
		// Funds:        fmt.Sprint(bal),
		Funds:        fmt.Sprint(*n.Amount),
		Ticketnumber: int(n.Ticket),
		Action:       []CommandType{n.Topic},
	}

	ulog, _ := json.Marshal(a)
	// bodyReader := bytes.NewReader(ulog)
	// _, err := http.Post("http://"+logHOST+":8004/accountlog", "application/json", bodyReader)
	err := sendLog("accountlog", ulog)
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
	// bodyReader := bytes.NewReader(ulog)
	// _, err := http.Post("http://"+logHOST+":8004/errorlog", "application/json", bodyReader)
	err := sendLog("errorlog", ulog)
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
		// bodyReader := bytes.NewReader(ulog)
		// log.Println(string(ulog))
		// _, err := http.Post("http://"+logHOST+":8004/debuglog", "application/json", bodyReader)
		err := sendLog("debuglog", ulog)
		if err != nil {
			log.Println(err)
		}

	}
}

type user_log struct {
	Username     UserId   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Command      []string `xml:"command,attr" json:"command"`
}

type system_log struct {
	Username     UserId   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Command      []string `xml:"command,attr" json:"command"`
}

type account_log struct {
	Username     UserId        `xml:"username" json:"username"`
	Funds        string        `xml:"funds" json:"funds"`
	Ticketnumber int           `xml:"ticketnumber" json:"ticketnumber"`
	Action       []CommandType `xml:"action,attr" json:"action"`
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
