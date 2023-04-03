package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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
	bodyReader := bytes.NewReader(ulog)
	_, err := http.Post("http://"+logHOST+":8004/userlog", "application/json", bodyReader)
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
	bodyReader := bytes.NewReader(ulog)
	_, err := http.Post("http://"+logHOST+":8004/accountlog", "application/json", bodyReader)
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
