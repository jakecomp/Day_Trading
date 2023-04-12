package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	for key, value := range known_stocks.stocks {

		const max = 500.0
		const min = 10.0

		rand_price := min + rand.Float64()*(max-min)
		value.Price = rand_price
		known_stocks.stocks[key] = value

	}

	c.IndentedJSON(http.StatusOK, known_stocks.stocks)
}

func main() {

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	gin.SetMode(gin.ReleaseMode)
	known_stocks.stocks = make(map[string]quote, 0)
	known_stocks.stocks["S"] = quote{"S", float64(100)}
	fmt.Println("Quote Server runing on Port 8002...")
	router := gin.Default()
	router.GET("/", gen_quote)
	// To get the stock price of R
	// curl localhost:8002/R
	router.GET("/:id", gen_quote_for_stock)
	// For all stocks
	// curl localhost:8002/all
	router.GET("/all", gen_quote_for_all_stocks)

	//router.Run("10.9.0.6:8002")
	//http.ListenAndServe("10.9.0.6:8002", router)

	srv := &http.Server{
		Addr:    ":8002",
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")

}
