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
	// fmt.Println(string(out))
	f.WriteString(string(out))
	f.WriteString("\n")
}
