package main

import (
	"sync"
	"time"
)

type Notification struct {
	Timestamp time.Time
	Topic     CommandType
	Ticket    int64
	Userid    UserId
	Stock     *string
	Amount    *float64
}

type MessageBus struct {
	subscriptions map[CommandType][]chan Notification
	backlog       map[CommandType]map[UserId]*Notification
	lock          sync.Mutex
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		subscriptions: make(map[CommandType][]chan Notification),
		backlog:       make(map[CommandType]map[UserId]*Notification),
	}
}

func (mb *MessageBus) SubscribeAll(topic CommandType) <-chan Notification {
	return mb.Subscribe(topic, "*ALL*")
}

func (mb *MessageBus) Subscribe(topic CommandType, uid UserId) chan Notification {
	mb.lock.Lock()
	defer mb.lock.Unlock()
	ch := make(chan Notification)
	mb.subscriptions[topic] = append(mb.subscriptions[topic], ch)

	// Notify this new subscription of the last broadcasted
	// notification on this topic
	if mb.backlog[topic] == nil {
		mb.backlog[topic] = make(map[UserId]*Notification)
	}
	msg := mb.backlog[topic][uid]
	go func() {
		if msg != nil {
			ch <- *msg
		}
	}()

	return ch
}

func (mb *MessageBus) Publish(topic CommandType, message Notification) {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	for _, ch := range mb.subscriptions[topic] {

		go func(c chan Notification) {
			c <- message
		}(ch)
	}

	// Backup the latest notification to update subscribers
	if mb.backlog[topic] == nil {
		mb.backlog[topic] = make(map[UserId]*Notification)
	}
	mb.backlog[topic][UserId(message.Userid)] = &message
}
