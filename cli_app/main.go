package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
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

func parseCmds(r *bufio.Reader) chan []Command {
	c := make(chan []Command, 500)
	usercmds := make(map[string][]Command, 1000)
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
			usercmds[cmd.Args[0]] = append(usercmds[cmd.Args[0]], *cmd)
		}
		for _, value := range usercmds {
			c <- value
		}
		close(c)
	}()
	return c
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type token []byte

func marshalAndSend(user User, requestUrl string) (body []byte) {
	usrInf, err := json.Marshal(user)
	bodyReader := bytes.NewReader(usrInf)

	if err != nil {
		log.Fatal("Failed to marshal ", err)
	}

	res, err := http.Post(requestUrl, "application/json", bodyReader)

	if err != nil {
		log.Fatal("Failed to connect ", err)
	}
	bod, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Failed to read body ", err)
	}
	return bod
}

func signup(user User) error {
	body := marshalAndSend(user, "http://localhost:8000/signup")

	fmt.Printf("res was %s\n", string(body))
	if strings.Compare(string(body), `SIGNED YOU UP!`) != 0 {
		return errors.New("Failed to sign up")
	}
	return nil

}

func signin(user User) (token, error) {
	body := marshalAndSend(user, "http://localhost:8000/signin")

	fmt.Printf("res was %s\n", string(body))
	return token(body), nil
}

func connectToSocket(tok token) *websocket.Conn {
	addr := "localhost:8000"
	u := "ws://" + addr + "/ws?token=" + string(tok)
	log.Printf("connecting to %s", u)

	// Connect to websocket server on localhost:8000/echo
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	return c
}

func forwardCommands(cmdsPerUser chan []Command, c *websocket.Conn) {
	for cmds := range cmdsPerUser {
		jcmd, err := json.Marshal(cmds)
		if err != nil {
			log.Println("failed during command marshelling ", err)
		}
		for _, cm := range cmds {
			log.Println(cm)
		}
		err = c.WriteMessage(websocket.TextMessage, jcmd)
		// mType, m, err := c.ReadMessage()
		if err != nil {
			fmt.Println("error on read message: ", err, cmds)
		}
		// if mType == websocket.TextMessage {
		// 	log.Println(string(m))
		// }
	}
	c.Close()

}

func main() {
	scanner := bufio.NewReader(os.Stdin)
	cmds := parseCmds(scanner)
	usr := User{Username: "testing", Password: "lol"}
	err := signup(usr)
	if err != nil {
		log.Fatal(err)
	}

	tok, err := signin(usr)

	if err != nil {
		log.Fatal(err)
	}

	c := connectToSocket(tok)
	defer c.Close()
	forwardCommands(cmds, c)
}
