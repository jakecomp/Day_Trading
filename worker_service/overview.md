# Structure of this project

## worker_service.go

This is the core functionality of the worker_service and likely the
file you want to modify unless you are adding a new command

The most **important functions** are as follows

- stockMonitor
  - subscribes to any BUY or SELL commands and will query the stock price from the quote server 
- UserAccountManager
  - Subscribes to `ADD`, `BUY_COMMIT`, and `SELL_COMMIT`. These are then
    used to execute each operation one at a time while still being non
    blocking.
- commandLogger
  - Simply logs all commands when a Notification is sent. Only really
    used for logging
- getNextCommand
  - Queries a new command from the queue service
- dispatch
  - When a new command as received dispatch will construct a struct
    associated with that command and return it as well as an error if
    necessary

getNextCommand and dispatch are used together to perform commands as
the queue service sends them.

## bus.go

This contains all the methods and functionality used for the message bus struct

Users of the message bus can Subscribe or Publish to the message bus with the methods

mb.SubscribeAll("BUY") 
mb.Subscribe("myuser" "BUY")
mb.Publish("BUY", Notification{ ... })

Subscribe will return a channel that recieves all Notifcations
published for the given topic.

So mb.Subscribe("myuser" "BUY") will return a channel listening for
notifications like mb.publish("BUY", Notification{ ... }) where the
Notification's Userid matches "myuser".

SubscribeAll does the same but does discriminate by user and will be sent all
commands related to that topic.

Finally we have Publish which is simply a way to send a notification
to any subscribers

### Example Usage

```go

func main() {
	go func() {
		ch := mb.Subscribe(notifyADD, userid("gavin"))
		n <-ch
		fmt.Println(n.Userid)
	}()
		

	mb.Publish(notifyADD, Notification{
		Topic:     notifyAdd,
		Timestamp: time.Now(),
		Ticket:    100,
		Userid:    "gavin",
		Stock:     nil,
		Amount:    200.0,
	})
	mb.Publish(notifyADD,  Notification{
		Topic:     notifyAdd,
		Timestamp: time.Now(),
		Ticket:    100,
		Userid:    "gavin",
		Stock:     nil,
		Amount:    200.0,
	})
	mb.Publish(notifyADD, Notification{
		Topic:     notifyAdd,
		Timestamp: time.Now(),
		Ticket:    100,
		Userid:    "Jim",
		Stock:     nil,
		Amount:    200.0,
	})
}
```

This would print "gavin" 2 times since the last notification does have the userid of "gavine" it is never sent to the subscription

All operations are thread safe since I used channels for all of this
so there isn't much need to worry about that

**IMPORTANT** A single back log of each Notification is stored by the MessageBus

This acts as a cache and avoids the need to ensure dependant commands are not missed

e.g. if a COMMIT_BUY occurs before a BUY and the BUY has an earlier
Ticket number then the message bus will still be able to access it.

```go
func main() {
	mb.Publish(notifyADD, Notification{
		Topic:     notifyAdd,
		Timestamp: time.Now(),
		Ticket:    100,
		Userid:    "gavin",
		Stock:     nil,
		Amount:    200.0,
	})

	go func() {
		ch := mb.Subscribe(notifyADD, userid("gavin"))
		n <-ch
		fmt.Println(n.Userid)
	}()
		

	mb.Publish(notifyADD,  Notification{
		Topic:     notifyAdd,
		Timestamp: time.Now(),
		Ticket:    100,
		Userid:    "gavin",
		Stock:     nil,
		Amount:    200.0,
	})
}
```

This will still print "gavin" twice since the first publication is
stored in the MessageBus's backlog

the structure of the MessageBus is as follows

```go
type MessageBus struct {
	subscriptions map[string][]chan Notification
	backlog       map[string]map[userid]*Notification
	lock          sync.Mutex
}
```

## commands.go

This file contains the describing of all commands supported by the worker.

Each command is a dedicated struct that follows the `CMD` interface

```
type CMD interface {
	Notify() Notification
	Prerequsite(*MessageBus) error
	Execute(ch chan *Transaction) error
	Postrequsite(*MessageBus) error
}
```

This interface is used by the `Run` to execute them in the following sequence
- `c.Prerequsite(m)`
- `c.Execute(tchan)`
- `go func() { n := c.Notify() m.Publish(n.Topic, n) }()`
- `c.Postrequsite(m)`

Note that the reason Notification is done as seen above is to avoid blocking during 
publications.

This allows commands to be self describing and easily conform to the
description given to us by relying on the MessageBus and having each
step in a command be dependant on existing subscriptions
