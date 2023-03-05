package main

import (
	"os"
	"testing"
)

func TestCommandsBuyCommit(t *testing.T) {
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(ADD{userId: "me", amount: 32.1}, mb)
	go Run(BUY{userId: "me", stock: "ABC", cost: 32.1, amount: 1.0}, mb)
	go func() {
		Run(&COMMIT_BUY{userId: "me"}, mb)
		finch <- nil
	}()
	err := <-finch
	if err != nil {
		t.Fail()
	}
}
func TestCommandsBuyCancel(t *testing.T) {
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(ADD{userId: "me", amount: 32.1}, mb)
	go Run(BUY{userId: "me", stock: "ABC", cost: 32.1, amount: 1.0}, mb)
	go func() {
		Run(&CANCEL_BUY{userId: "me"}, mb)
		finch <- nil
	}()
	err := <-finch
	if err != nil {
		t.Fail()
	}
}

func TestCommandsSELLCommit(t *testing.T) {
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(SELL{userId: "me", stock: "ABC", cost: 32.1, amount: 1.0}, mb)
	go func() {
		Run(&COMMIT_SELL{userId: "me"}, mb)
		finch <- nil
	}()
	err := <-finch
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
