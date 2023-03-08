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
	// "github.com/golang-jwt/jwt/v4"
	// "github.com/gorilla/websocket"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	// "golang.org/x/crypto/bcrypt"
	//"github.com/shabbyrobe/xmlwriter"
)

type userCommand struct {
	XmlName      xml.Name `xml:"userCommand"`
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Command      []string `xml:"command" json:"command"`
}

type accountTransaction struct {
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Action       []string `xml:"action" json:"action"`
}

type quoteServer struct {
	Timestamp       int64  `xml:"timestamp"`
	ServerName      string `xml:"server" json:"server"`
	Username        string `xml:"username" json:"username"`
	Ticketnumber    int    `xml:"ticketnumber" json:"ticketnumber"`
	Price           string `xml:"price" json:"price"`
	QuoteServerTime int64  `xml:"quoteServerTime" json:"quoteServerTime"`
	StockSymbol     string `xml:"stock_symbol" json:"stock_symbol"`
}

type systemEvent struct {
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Username     string   `xml:"username" json:"username"`
	Funds        string   `xml:"funds" json:"funds"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Command      []string `xml:"command,attr" json:"command"`
}

type errorEvent struct {
	Timestamp    int64    `xml:"timestamp"`
	ServerName   string   `xml:"server" json:"server"`
	Username     string   `xml:"username" json:"username"`
	Ticketnumber int      `xml:"ticketnumber" json:"ticketnumber"`
	Command      []string `xml:"command" json:"command"`
}

var f *os.File

//var ec *xmlwriter.ErrCollector
// var b  *bytes.Buffer
// var ec ErrCollector

func main() {
	var err error

	_, err = os.Stat("stocklog.txt")

	if !errors.Is(err, os.ErrNotExist) {

		os.Remove("stocklog.txt")
	}

	f, err = os.OpenFile("stocklog.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {

		log.Fatalf("error opening file: %v", err)

	}

	fmt.Println("Running on Port number: 8004")

	//	errorlog("hello")
	// recive_log := &user_log{
	// 	//XmlName:      xml.Name{},
	// 	Timestamp:    0,
	// 	Username:     "Dog",
	// 	Funds:        "1432",
	// 	Ticketnumber: 80,
	// 	Command:      []string{},
	// }
	// //recive_log.XmlName.Local = "UserCommand"
	// recive_log.Timestamp = timestamp()
	// out, _ := xml.MarshalIndent(recive_log, "", "\t")
	// fmt.Println(string(out))
	// f.WriteString(string(out))
	// f.WriteString("\n")
	handleRequests()
	//fmt.Println("'yes'")
	f.Close()

}

func timestamp() int64 {

	return time.Now().Unix()

}

func handleRequests() {

	http.HandleFunc("/userlog", userlog)

	http.HandleFunc("/accountlog", accountlog)

	http.HandleFunc("/quotelog", quotelog)

	http.HandleFunc("/systemlog", systemlog)

	http.HandleFunc("/errorlog", errorlog)

	log.Fatal(http.ListenAndServe(":8004", nil))

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
	//_ = xml.NewEncoder(*xmlwriter).Encode(recive_log)
	out, _ := xml.MarshalIndent(recive_log, "", "\t")
	//fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")
}
