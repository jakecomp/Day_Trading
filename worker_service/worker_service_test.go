package main

import (
	"context"
	"errors"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"os"
	"testing"
)

type MockPendingTransaction struct {
	pending map[string]*Notification
}

func (m *MockPendingTransaction) lastPending(uid UserId, topic CommandType) (*Notification, error) {
	n, ok := m.pending[string(uid)+"#"+string(topic)]
	if !ok {
		return nil, errors.New("failed to find")
	}
	delete(m.pending, string(uid)+"#"+string(topic))
	return n, nil
}

func (m *MockPendingTransaction) Store(uid UserId, topic CommandType, n *Notification) error {
	m.pending[string(uid)+"#"+string(topic)] = n
	return nil
}

type MockUserTransactor struct {
	users map[UserId]*user_doc
}

func (m *MockUserTransactor) Execute(t func(context.Context) error) error {
	return t(context.Background())
}

func (m *MockUserTransactor) getUser(c CommandType, uid UserId) (*user_doc, error) {
	n, ok := m.users[uid]
	if !ok {
		if c == notifyADD {
			var new_doc = new(user_doc)
			new_doc.Username = uid
			new_doc.Hash = "unsecure_this_user_never_made_account_via_backend"
			new_doc.Balance = 0
			new_doc.Stonks = make(map[string]float64)
			return new_doc, nil

		}
		return nil, errors.New("failed to find")
	}
	return n, nil
}
func (m *MockUserTransactor) setUser(username UserId, balance float32, stocks map[string]float64) error {
	log.Println("setting user", username, balance, stocks)
	m.users[username] = &user_doc{
		Username: username,
		Balance:  balance,
		Stonks:   stocks,
	}
	return nil

}

type MockStockPriceSource struct {
	stocks map[string]Stock
}

func (m *MockStockPriceSource) setPrice(stock string, price float64) error {
	m.stocks[stock] = Stock{
		Name:  stock,
		Price: price,
	}
	return nil
}

func (m *MockStockPriceSource) lookupPrice(stock string, ticket int64) (Stock, error) {
	const max = 500.0
	const min = 10.0

	n, ok := m.stocks[stock]
	if !ok {

		log.Println("making it up as I go along")
		m.stocks[stock] = Stock{
			Name:  stock,
			Price: rand.Float64() * (max - min),
		}
		return m.stocks[stock], errors.New("failed to find")
	}
	return n, nil
}

func FakeRun(commands []Command, stock_prices map[string]Stock) (*MockPendingTransaction, *MockUserTransactor, *MockStockPriceSource, error) {
	isTesting = true
	mb := NewMessageBus()
	pendingTransactions := &MockPendingTransaction{
		pending: make(map[string]*Notification),
	}
	users := &MockUserTransactor{
		users: make(map[UserId]*user_doc),
	}
	stock_pricer := &MockStockPriceSource{
		stocks: stock_prices,
	}

	for _, c := range commands {
		cmd, err := dispatch(c)
		if err == nil {
			// Execute this new command
			Run(cmd, mb, pendingTransactions, users, stock_pricer)
		} else {
			return nil, nil, nil, err
		}
	}
	return pendingTransactions, users, stock_pricer, nil
}

func TestAdd(t *testing.T) {
	_, u, _, err := FakeRun(
		[]Command{{Ticket: 1, Command: "ADD", Args: Args{"umhaEY4lil", "91628.00"}}},
		map[string]Stock{
			"umhaEY4lil": {Name: "umhaEY4lil", Price: 10},
		})
	usr, err := u.getUser(notifyADD, "umhaEY4lil")
	if err != nil {
		t.Errorf("cant even find user")
	}
	if usr.Username != "umhaEY4lil" || usr.Balance != 91628.00 {
		t.Error("user ,", usr, " is not expected value %", "umhaEY4lil")
	}
}

func TestBuy(t *testing.T) {
	uid := "umhaEY4lil"
	pen, u, _, err := FakeRun(
		[]Command{{Ticket: 1, Command: "ADD", Args: Args{uid, "91628.00"}},
			{Ticket: 2, Command: "BUY", Args: Args{uid, "HWU", "91628.00"}}},
		map[string]Stock{
			uid: {Name: uid, Price: 1},
		})
	usr, err := u.getUser(notifyADD, UserId(uid))
	if err != nil {
		t.Errorf("cant even find user")
	}
	pending, err := pen.lastPending(UserId(uid), notifyBUY)
	if usr.Username != UserId("umhaEY4lil") || usr.Balance != 91628.00 {
		t.Error("user ,", usr, " is not expected value %", "umhaEY4lil")
	}
	if err != nil || *pending.Stock != "HWU" || pending.Userid != usr.Username || *pending.Amount != 91628.00 {
		t.Error("user ,", usr, " is not expected value %", "umhaEY4lil")

	}
}

func TestCommitBuy(t *testing.T) {
	uid := "umhaEY4lil"
	stock := "HWU"
	pen, u, _, err := FakeRun(
		[]Command{{Ticket: 1, Command: "ADD", Args: Args{uid, "91628.00"}},
			{Ticket: 2, Command: "BUY", Args: Args{uid, stock, "91628.00"}},
			{Ticket: 3, Command: notifyCOMMIT_BUY, Args: Args{uid, stock}}},
		map[string]Stock{
			stock: {Name: stock, Price: 1},
		})
	usr, err := u.getUser(notifyADD, UserId(uid))
	if err != nil {
		t.Errorf("cant even find user")
	}
	pending, err := pen.lastPending(UserId(uid), notifyBUY)
	if pending != nil {
		t.Error("stock still pending ", pending)
	}
	commited, ok := usr.Stonks[stock]
	if usr.Username != UserId("umhaEY4lil") || usr.Balance != 0 || !ok || commited != 91628 {
		t.Error("user ,", usr, " is not expected value ", uid, " ", commited)
	}
}
func TestSell(t *testing.T) {
	uid := "umhaEY4lil"
	stock := "HWU"
	pen, u, _, err := FakeRun(
		[]Command{{Ticket: 1, Command: "ADD", Args: Args{uid, "91628.00"}},
			{Ticket: 2, Command: "BUY", Args: Args{uid, stock, "91628.00"}},
			{Ticket: 3, Command: notifyCOMMIT_BUY, Args: Args{uid}},
			{Ticket: 4, Command: notifySELL, Args: Args{uid, stock, "91628.00"}},
			{Ticket: 5, Command: notifyCOMMIT_SELL, Args: Args{uid}}},
		map[string]Stock{
			stock: {Name: stock, Price: 1},
		})
	usr, err := u.getUser(notifyADD, UserId(uid))
	if err != nil {
		t.Errorf("cant even find user")
	}
	pending, err := pen.lastPending(UserId(uid), notifyBUY)
	if pending != nil {
		t.Error("stock still pending ", pending)
	}
	commited, ok := usr.Stonks[stock]
	if usr.Username != UserId("umhaEY4lil") || usr.Balance != 91628 || !ok || commited != 0 {
		t.Error("user ,", usr, " is not expected value ", uid, " ", commited)
	}
}

func TestSellCancel(t *testing.T) {
	uid := "umhaEY4lil"
	stock := "HWU"
	pen, u, _, err := FakeRun(
		[]Command{{Ticket: 1, Command: "ADD", Args: Args{uid, "91628.00"}},
			{Ticket: 2, Command: "BUY", Args: Args{uid, stock, "91628.00"}},
			{Ticket: 3, Command: notifyCOMMIT_BUY, Args: Args{uid}},
			{Ticket: 4, Command: notifySELL, Args: Args{uid, stock, "91628.00"}},
			{Ticket: 5, Command: notifyCANCEL_SELL, Args: Args{uid}}},
		map[string]Stock{
			stock: {Name: stock, Price: 1},
		})
	usr, err := u.getUser(notifyADD, UserId(uid))
	if err != nil {
		t.Errorf("cant even find user")
	}
	pending, err := pen.lastPending(UserId(uid), notifyBUY)
	if pending != nil {
		t.Error("stock still pending ", pending)
	}
	commited, ok := usr.Stonks[stock]
	if usr.Username != UserId("umhaEY4lil") || usr.Balance != 0 || !ok || commited != 91628 {
		t.Error("user ,", usr, " is not expected value ", uid, " ", commited)
	}
}

func TestBuyCancel(t *testing.T) {
	uid := "umhaEY4lil"
	stock := "HWU"
	pen, u, _, err := FakeRun(
		[]Command{{Ticket: 1, Command: "ADD", Args: Args{uid, "91628.00"}},
			{Ticket: 2, Command: "BUY", Args: Args{uid, stock, "91628.00"}},
			{Ticket: 5, Command: notifyCANCEL_BUY, Args: Args{uid}}},
		map[string]Stock{
			stock: {Name: stock, Price: 1},
		})
	usr, err := u.getUser(notifyADD, UserId(uid))
	if err != nil {
		t.Errorf("cant even find user")
	}
	pending, err := pen.lastPending(UserId(uid), notifyBUY)
	if pending != nil {
		t.Error("stock still pending ", pending)
	}
	commited, ok := usr.Stonks[stock]
	if usr.Username != UserId("umhaEY4lil") || usr.Balance != 91628.00 || ok {
		t.Error("user ,", usr, " is not expected value ", uid, " ", commited)
	}
}

// TODO determine how to catch errors being logged
func TestFailedBuyCancel(t *testing.T) {
	uid := "umhaEY4lil"
	stock := "HWU"
	pen, u, _, err := FakeRun(
		[]Command{{Ticket: 1, Command: "ADD", Args: Args{uid, "91628.00"}},
			{Ticket: 5, Command: notifyCANCEL_BUY, Args: Args{uid}}},
		map[string]Stock{
			stock: {Name: stock, Price: 1},
		})
	usr, err := u.getUser(notifyADD, UserId(uid))
	if err != nil {
		t.Errorf("cant even find user")
	}
	pending, err := pen.lastPending(UserId(uid), notifyBUY)
	if pending != nil {
		t.Error("stock still pending ", pending)
	}
	commited, ok := usr.Stonks[stock]
	if usr.Username != UserId("umhaEY4lil") || usr.Balance != 91628.00 || ok {
		t.Error("user ,", usr, " is not expected value ", uid, " ", commited)
	}
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

// func TestSetupUserTasks(t *testing.T) {
// 	conn, err := dial("amqp://guest:guest@" + rabbitmqHOST + ":5672/")
// 	// connect to an existing exchange
// 	userid := "Xyz"
// 	userid2 := "Xyzl"

// 	print("setup worker")
// 	setupNewUserInBackend(userid, conn).Publish("", userid, false, false,
// 		amqp.Publishing{
// 			ContentType: "text/plain",
// 			Body:        []byte("hello"),
// 		},
// 	)
// 	setupNewUserInBackend(userid2, conn).Publish("", userid2, false, false,
// 		amqp.Publishing{
// 			ContentType: "text/plain",
// 			Body:        []byte("hello"),
// 		},
// 	)
// 	fmt.Println("setup backend")
// 	done1 := make(chan struct{})
// 	c, _ := getNextMessageForDictator(conn)
// 	dictator := func() {
// 		for {
// 			l := <-c
// 			fmt.Println("got value", l.RoutingKey)
// 			userid := l.RoutingKey
// 			fmt.Println("routing key was ", userid)
// 			conchan, _ := conn.Channel()
// 			q, _ := conchan.QueueInspect(userid)
// 			fmt.Println("queue found ", q)
// 			if q.Consumers == 0 {
// 				c, err := conchan.Consume(q.Name, "", false, false, false, false, nil)
// 				cmd := <-c
// 				fmt.Println(string(cmd.Body), err)
// 				conchan.Close()
// 				// done1 <- struct{}{}
// 			}

// 		}
// 	}
// 	dictator()
// 	<-done1
// 	<-done1

// 	if err != nil {

// 		t.Fail()
// 	}
// }

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
