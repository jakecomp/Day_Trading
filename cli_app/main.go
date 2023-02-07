package main

import (
	"bufio"
	"bytes"
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
	ticket  int
	command string
	args    []string
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
	return &Command{ticket: ticketnum, command: parts[0], args: parts[1:]}, nil
}

func parseCmds(r *bufio.Reader) chan Command {
	c := make(chan Command)
	go func() {
		for l, _, err := r.ReadLine(); err == nil; l, _, err = r.ReadLine() {
			fmt.Println(string(l))
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

func main() {
	requestUrl := "http://localhost:8000/signin"
	bodyReader := bytes.NewReader([]byte(`{ "username": "testing", "password": "lol"}`))
	res, err := http.Post(requestUrl, "application/json", bodyReader)
	if err != nil {
		fmt.Printf("Failed to connect %s\n", err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Printf("Failed to read body %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("res was %d", body)
	// scanner := bufio.NewReader(os.Stdin)
	// // TODO establish connection
	// for cmd := range parseCmds(scanner) {
	// 	fmt.Println(cmd)
	// }
	// os.Exit(1)
	fmt.Printf("connecting\n")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var addr string = "localhost:8000"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	// Connect to websocket server on localhost:8000/echo
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	c.WriteMessage(websocket.TextMessage, []byte("hello"))
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	fmt.Println("Hello world")
}
