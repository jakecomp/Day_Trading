package main

import (
	"fmt"
	"math/rand"
	"net/http"

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

	var new_quote = []quote{

		{Stock: "s", Price: rand_price},
	}

	c.IndentedJSON(http.StatusOK, new_quote)

}

func main() {

	fmt.Println("Quote Server runing on Port 8002...")
	router := gin.Default()
	router.GET("/", gen_quote)

	router.Run("10.9.0.6:8002")
}
