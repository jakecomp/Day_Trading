package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type quote struct {
	Stock string  `json:"stock"`
	Price float64 `json:"price"`
}

func gen_quote(c *gin.Context) {

	var max = 500.0
	var min = 10.0

	var rand_price = min + rand.Float64()*(max-min)

	var new_quote = quote{
		Stock: "S", Price: rand_price,
	}

	c.IndentedJSON(http.StatusOK, new_quote)

}

type KnownStocks struct {
	stocks map[string]quote
	lock   sync.Mutex
}

var known_stocks KnownStocks

func gen_quote_for_stock(c *gin.Context) {
	known_stocks.lock.Lock()
	defer known_stocks.lock.Unlock()

	id := c.Param("id")

	const max = 500.0
	const min = 10.0

	rand_price := min + rand.Float64()*(max-min)

	new_quote := quote{
		Stock: id, Price: rand_price,
	}
	known_stocks.stocks[id] = new_quote
	c.IndentedJSON(http.StatusOK, new_quote)

}
func gen_quote_for_all_stocks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, known_stocks.stocks)
}

func main() {
	known_stocks.stocks = make(map[string]quote, 0)
	fmt.Println("Quote Server runing on Port 8002...")
	router := gin.Default()
	router.GET("/", gen_quote)
	// To get the stock price of R
	// curl localhost:8002/R
	router.GET("/:id", gen_quote_for_stock)
	// curl localhost:8002/all
	router.GET("/all", gen_quote_for_all_stocks)

	router.Run("10.9.0.6:8002")
}
