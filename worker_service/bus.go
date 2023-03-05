package main

import (
	"sync"
	"time"
)

type Notification struct {
	Timestamp time.Time
	Userid    string
	Stock     *string
	Amount    *float64
}

type MessageBus struct {
	subscriptions map[string][]chan Notification
	backlog       map[string]map[userid]*Notification
	lock          sync.Mutex
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		subscriptions: make(map[string][]chan Notification),
		backlog:       make(map[string]map[userid]*Notification),
	}
}

func (mb *MessageBus) Subscribe(topic string, uid userid) chan Notification {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	ch := make(chan Notification)
	mb.subscriptions[topic] = append(mb.subscriptions[topic], ch)

	if mb.backlog[topic] == nil {
		mb.backlog[topic] = make(map[userid]*Notification)
	}
	msg := mb.backlog[topic][uid]
	go func() {
		if msg != nil {
			ch <- *msg
		}
	}()

	return ch
}

func (mb *MessageBus) Publish(topic string, message Notification) {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	for _, ch := range mb.subscriptions[topic] {
		go func(c chan Notification) {
			c <- message
		}(ch)
	}

	if mb.backlog[topic] == nil {
		mb.backlog[topic] = make(map[userid]*Notification)
	}
	mb.backlog[topic][userid(message.Userid)] = &message
}
