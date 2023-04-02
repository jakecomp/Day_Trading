package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"os"
	"testing"
)

func setupNewUserInBackend(userid string, conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ")
	// err = ch.ExchangeDeclarePassive(
	// 	"user_tasks", // name
	// 	"topic",      // type
	// 	true,         // durable
	// 	false,        // auto-deleted
	// 	false,        // internal
	// 	false,        // no-wait
	// 	nil,          // arguments
	// )
	// // If it does not ex ist
	// if err != nil {
	print("creating exchange")
	err = ch.ExchangeDeclare(
		"user_tasks", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	// }
	q, err := ch.QueueDeclare(
		userid, // name
		true,   // durable
		false,  // delete when unused
		true,   // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	err = ch.QueueBind(
		q.Name,       // queue name
		userid,       // routing key
		"user_tasks", // exchange
		false,
		nil)
	// failOnError(err, "Failed to bind to exchange")

	return ch
}

// func setupForinternalWorker(user string, conn *amqp.Connection) (<-chan amqp.Delivery, error) {
// 	conn, err := dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
// 	ch, err := conn.Channel()
// 	failOnError(err, "Failed to connect to RabbitMQ")

// 	err = ch.ExchangeDeclarePassive(
// 		"user_tasks", // name
// 		"topic",      // type
// 		true,         // durable
// 		false,        // auto-deleted
// 		false,        // internal
// 		false,        // no-wait
// 		nil,          // arguments
// 	)
// 	// If it does not exist
// 	if err != nil {
// 		err = ch.ExchangeDeclare(
// 			"user_tasks", // name
// 			"topic",      // type
// 			true,         // durable
// 			false,        // auto-deleted
// 			false,        // internal
// 			false,        // no-wait
// 			nil,          // arguments
// 		)
// 	}

// 	fromBackend := "user_tasks#worker"
// 	// q, err := ch.QueueDeclarePassive(
// 	// 	fromBackend, // name
// 	// 	false,       // durable
// 	// 	false,       // delete when unused
// 	// 	true,        // exclusive
// 	// 	false,       // no-wait
// 	// 	nil,         // arguments
// 	// )
// 	// // If it does not exist
// 	// if err != nil {
// 	q, err := ch.QueueDeclare(
// 		fromBackend, // name
// 		false,       // durable
// 		false,       // delete when unused
// 		false,       // exclusive
// 		false,       // no-wait
// 		nil,         // arguments
// 	)
// 	// }
// 	err = ch.QueueBind(
// 		q.Name,       // queue name
// 		user,         // routing key
// 		"user_tasks", // exchange
// 		false,
// 		nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return ch.Consume(q.Name, "", false, false, false, false, nil)

// }

// func setupForWorker(conn *amqp.Connection) (<-chan amqp.Delivery, error) {
// 	ch, err := conn.Channel()
// 	failOnError(err, "Failed to connect to RabbitMQ")

// 	err = ch.ExchangeDeclarePassive(
// 		"user_tasks", // name
// 		"topic",      // type
// 		true,         // durable
// 		false,        // auto-deleted
// 		false,        // internal
// 		false,        // no-wait
// 		nil,          // arguments
// 	)
// 	// If it does not exist
// 	if err != nil {
// 		err = ch.ExchangeDeclare(
// 			"user_tasks", // name
// 			"topic",      // type
// 			true,         // durable
// 			false,        // auto-deleted
// 			false,        // internal
// 			false,        // no-wait
// 			nil,          // arguments
// 		)
// 	}

//		fromBackend := "user_tasks#worker"
//		// q, err := ch.QueueDeclarePassive(
//		// 	fromBackend, // name
//		// 	false,       // durable
//		// 	false,       // delete when unused
//		// 	true,        // exclusive
//		// 	false,       // no-wait
//		// 	nil,         // arguments
//		// )
//		// // If it does not exist
//		// if err != nil {
//		q, err := ch.QueueDeclare(
//			fromBackend, // name
//			false,       // durable
//			false,       // delete when unused
//			true,        // exclusive
//			false,       // no-wait
//			nil,         // arguments
//		)
//		// }
//		err = ch.QueueBind(
//			q.Name,       // queue name
//			"*",          // routing key
//			"user_tasks", // exchange
//			false,
//			nil)
//		if err != nil {
//			panic(err)
//		}
//		return ch.Consume(q.Name, "", false, false, false, false, nil)
//	}
func getNextMessageForDictator(conn *amqp.Connection) (<-chan amqp.Delivery, error) {
	ch, err := conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ")

	// err = ch.ExchangeDeclarePassive(
	// 	"user_tasks", // name
	// 	"topic",      // type
	// 	true,         // durable
	// 	false,        // auto-deleted
	// 	false,        // internal
	// 	false,        // no-wait
	// 	nil,          // arguments
	// )
	// // If it does not exist
	// if err != nil {
	err = ch.ExchangeDeclare(
		"user_tasks", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	// }

	fromBackend := "dick"
	// q, err := ch.QueueDeclarePassive(
	// 	fromBackend, // name
	// 	false,       // durable
	// 	false,       // delete when unused
	// 	true,        // exclusive
	// 	false,       // no-wait
	// 	nil,         // arguments
	// )
	// // If it does not exist
	// if err != nil {
	q, err := ch.QueueDeclare(
		fromBackend, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive TODO make this true
		false,       // no-wait
		nil,         // arguments
	)
	// }
	err = ch.QueueBind(
		q.Name,       // queue name
		"*",          // routing key
		"user_tasks", // exchange
		false,
		nil)
	if err != nil {
		panic(err)
	}
	return ch.Consume(q.Name, "", false, false, false, false, nil)
}

func TestSetupUserTasks(t *testing.T) {
	conn, err := dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
	// connect to an existing exchange
	userid := "Xyz"
	userid2 := "Xyzl"

	print("setup worker")
	setupNewUserInBackend(userid, conn).Publish("", userid, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("hello"),
		},
	)
	setupNewUserInBackend(userid2, conn).Publish("", userid2, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("hello"),
		},
	)
	fmt.Println("setup backend")
	done1 := make(chan struct{})
	c, _ := getNextMessageForDictator(conn)
	dictator := func() {
		for {
			l := <-c
			fmt.Println("got value", l.RoutingKey)
			userid := l.RoutingKey
			fmt.Println("routing key was ", userid)
			conchan, _ := conn.Channel()
			q, _ := conchan.QueueInspect(userid)
			fmt.Println("queue found ", q)
			if q.Consumers == 0 {
				c, err := conchan.Consume(q.Name, "", false, false, false, false, nil)
				cmd := <-c
				fmt.Println(string(cmd.Body), err)
				conchan.Close()
				// done1 <- struct{}{}
			}

		}
	}
	dictator()
	<-done1
	<-done1

	if err != nil {

		t.Fail()
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	if ret == 0 {
		teardown()
	}
	os.Exit(ret)
}

func setup() {

}

func teardown() {

}
