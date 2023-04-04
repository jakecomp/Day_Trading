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
	lock          sync.RWMutex
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		subscriptions: make(map[CommandType][]chan Notification),
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
	return ch
}

func (mb *MessageBus) Publish(topic CommandType, message Notification) {
	mb.lock.RLock()
	defer mb.lock.RUnlock()

	for _, ch := range mb.subscriptions[topic] {

		go func(c chan Notification) {
			c <- message
		}(ch)
	}
}
