package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
)

type Command struct {
	Ticket  int
	Command string
	Args    []string
}

func client() {
	var addr string = "localhost:8000"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	// Connect to websocket server on localhost:8000/echo
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
}

func parseCmd(line string) (*Command, error) {
	var ticketnum int
	var cmd string
	_, err := fmt.Sscanf(line, "[%d] %s", &ticketnum, &cmd)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(cmd, ",")
	return &Command{Ticket: ticketnum, Command: parts[0], Args: parts[1:]}, nil
}

func parseCmds(r *bufio.Reader) chan Command {
	c := make(chan Command)
	go func() {
		for l, _, err := r.ReadLine(); err == nil; l, _, err = r.ReadLine() {
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading input:", err)
			}

			// todo handle parsing error
			cmd, err := parseCmd(string(l))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error parsing input:", err)
			}
			c <- *cmd
		}
	}()
	return c
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type token []byte

func signup(user User) error {
	requestUrl := "http://localhost:8000/signup"

	usrInf, err := json.Marshal(user)
	fmt.Printf(string(usrInf))
	bodyReader := bytes.NewReader(usrInf)
	if err != nil {
		log.Fatal("Failed to Marshal", err)
	}

	res, err := http.Post(requestUrl, "application/json", bodyReader)
	if err != nil {
		log.Fatal("Failed to connect", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Failed to read body", err)
	}

	fmt.Printf("res was %s\n", string(body))
	if strings.Compare(string(body), `SIGNED YOU UP!`) != 0 {
		return err
	}
	return nil

}
func signin(user User) (token, error) {
	requestUrl := "http://localhost:8000/signin"
	usrInf, err := json.Marshal(user)
	bodyReader := bytes.NewReader(usrInf)

	if err != nil {
		log.Fatal("Failed to marshal ", err)
	}

	res, err := http.Post(requestUrl, "application/json", bodyReader)

	if err != nil {
		log.Fatal("Failed to connect ", err)
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatal("Failed to read body ", err)
	}

	fmt.Printf("res was %s\n", string(body))
	fmt.Printf("connecting\n")
	return token(body), nil
}

func main() {
	usr := User{Username: "testing", Password: "lol"}
	err := signup(usr)
	if err != nil {
		log.Fatal(err)
	}

	tok, err := signin(usr)

	if err != nil {
		log.Fatal(err)
	}

	interrupt := make(chan os.Signal, 1)

	signal.Notify(interrupt, os.Interrupt)

	addr := "localhost:8000"
	u := "ws://" + addr + "/ws?token=" + string(tok)
	log.Printf("connecting to %s", u)

	// Connect to websocket server on localhost:8000/echo
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	defer c.Close()

	if err != nil {
		log.Fatal("dial:", err)
	}

	mType, b, err := c.ReadMessage()

	if err != nil {
		log.Fatal("read: ", err)
	}

	if mType == websocket.TextMessage {
		log.Println(string(b))
	}

	scanner := bufio.NewReader(os.Stdin)
	// TODO establish connection
	for cmd := range parseCmds(scanner) {
		jcmd, err := json.Marshal(cmd)
		if err != nil {
			log.Println("failed during command marshelling ", err)
		}
		err = c.WriteMessage(websocket.TextMessage, jcmd)
		mType, m, err := c.ReadMessage()
		if err != nil {
			log.Fatal("error on read message: ", err, cmd)
		}
		if mType == websocket.TextMessage {
			log.Println(string(m))
		}
	}
	os.Exit(1)

}
