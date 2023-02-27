package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
var transactionQueue = make([]int, 0)

func enqueue(queue []int, element int) []int {
	queue = append(queue, element)
	fmt.Println("Enqueued:", element)
	return queue
}

func dequeue(queue []int) (int, []int) {
	element := queue[0]
	if len(queue) == 1 {
		var tmp = []int{}
		return element, tmp
	}

	return element, queue[1:]
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate User
	fmt.Println("Endpoint Hit: WS")

	// Upgrade our raw HTTP connection to a websocket based one
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()
	err = conn.WriteMessage(1, []byte("Its Queue Time"))
	socketReader(conn)
}

func socketReader(conn *websocket.Conn) {
	// Event Loop, Handle Comms in here
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error during message reading:", err)
			break
		}
		fmt.Printf("Received: %s", string(message))

		err = conn.WriteMessage(messageType, message)
		if err != nil {
			fmt.Println("Error during message writing:", err)
			break
		}
	}
}

func handleRequests() {
	http.HandleFunc("/ws", socketHandler)
	log.Fatal(http.ListenAndServe(":8001", nil))
}

func main() {
	fmt.Println("Queue Service Starting... Port 8001")
}
