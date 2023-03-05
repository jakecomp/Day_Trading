package main

import (
	"os"
	"testing"
)

func TestCommandsBuyCommit(t *testing.T) {
	ch := make(chan *Transaction)
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(ADD{userId: "me", amount: 32.1}, mb, ch)
	go Run(BUY{userId: "me", stock: "ABC", cost: 32.1, amount: 1.0}, mb, ch)
	go func() {
		Run(&COMMIT_BUY{userId: "me"}, mb, ch)
		finch <- nil
	}()

	addT := <-ch
	if addT.Command != notifyADD {
		t.Fatalf("This transaction should have been an ADD %v", addT.Command)
	}
	buyT := <-ch
	if buyT.Command != notifyCOMMIT_BUY {
		t.Fatalf("This transaction should have been a COMMIT_BUY but was %v", buyT.Command)
	}
	err := <-finch
	if err != nil {
		t.Fail()
	}
}

func TestCommandsBuyCancel(t *testing.T) {
	ch := make(chan *Transaction)
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(ADD{userId: "me", amount: 32.1}, mb, ch)
	go Run(BUY{userId: "me", stock: "ABC", cost: 32.1, amount: 1.0}, mb, ch)
	go func() {
		Run(&CANCEL_BUY{userId: "me"}, mb, ch)
		finch <- nil
	}()
	addT := <-ch
	if addT.Command != notifyADD {
		t.Fatalf("This transaction should have been an ADD %v", addT.Command)
	}
}

func TestCommandsSELLCommit(t *testing.T) {
	ch := make(chan *Transaction)
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(ADD{userId: "me", amount: 32.1}, mb, ch)
	go Run(BUY{userId: "me", stock: "ABC", cost: 32.1, amount: 1.0}, mb, ch)
	go Run(&COMMIT_BUY{userId: "me"}, mb, ch)
	go Run(SELL{userId: "me", stock: "ABC", cost: 33.1, amount: 1.0}, mb, ch)
	go func() {
		Run(&COMMIT_SELL{userId: "me"}, mb, ch)
		finch <- nil
	}()
	sell := []*Transaction{<-ch, <-ch, <-ch}
	if sell[0].Command != notifyADD {
		t.Fatalf("This transaction should have been an ADD %v", sell[0].Command)
	}
	if sell[1].Command != notifyCOMMIT_BUY {
		t.Fatalf("This transaction should have been an COMMIT_BUY %v", sell[1].Command)
	}
	if sell[2].Command != notifyCOMMIT_SELL {
		t.Fatalf("This transaction should have been an COMMIT_SELL %v", sell[2].Command)
	}
	err := <-finch
	if err != nil {
		t.Fail()
	}
}

func TestCommandsSELLCommitMultiUser(t *testing.T) {
	ch := make(chan *Transaction)
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(ADD{userId: "me", amount: 32.1}, mb, ch)
	go Run(BUY{userId: "me", stock: "ABC", cost: 32.1, amount: 1.0}, mb, ch)
	go Run(&COMMIT_BUY{userId: "me"}, mb, ch)
	go Run(SELL{userId: "me", stock: "ABC", cost: 33.1, amount: 1.0}, mb, ch)
	go Run(SELL{userId: "you", stock: "ABC", cost: 33.1, amount: 1.0}, mb, ch)
	go func() {
		Run(&COMMIT_SELL{userId: "me"}, mb, ch)
		finch <- nil
	}()
	sell := []*Transaction{<-ch, <-ch, <-ch}
	if sell[0].Command != notifyADD {
		t.Fatalf("This transaction should have been an ADD %v", sell[0].Command)
	}
	if sell[1].Command != notifyCOMMIT_BUY {
		t.Fatalf("This transaction should have been an COMMIT_BUY %v", sell[1].Command)
	}
	if sell[2].Command != notifyCOMMIT_SELL {
		t.Fatalf("This transaction should have been an COMMIT_SELL %v", sell[2].Command)
	}
	err := <-finch
	if err != nil {
		t.Fail()
	}
}

func TestCommandsBUYCommitMultiUser(t *testing.T) {
	ch := make(chan *Transaction)
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(ADD{userId: "me", amount: 35.0}, mb, ch)
	go Run(ADD{userId: "you", amount: 35.0}, mb, ch)
	go Run(BUY{userId: "me", stock: "ABC", cost: 33.1, amount: 1.0}, mb, ch)
	go func() {
		Run(&COMMIT_BUY{userId: "me"}, mb, ch)
		finch <- nil
	}()
	sellT := <-ch
	if sellT.Command != notifyADD && sellT.User_id != "you" && sellT.User_id != "me" {
		t.Fatalf("This transaction should have been an COMMIT_SELL %v", sellT.Command)
	}
	sellT2 := <-ch
	if sellT2.Command != notifyADD && sellT2.User_id != "you" && sellT2.User_id != "me" {
		t.Fatalf("This transaction should have been an COMMIT_SELL %v", sellT2.Command)
	}
	sellT3 := <-ch
	if sellT3.Command != notifyCOMMIT_BUY && sellT3.User_id != "me" {
		t.Fatalf("This transaction should have been an COMMIT_SELL %v", sellT2.Command)
	}

	err := <-finch
	if err != nil {
		t.Fail()
	}
}

func TestCommandsBUYNotEnoughMoney(t *testing.T) {
	ch := make(chan *Transaction)
	mb := NewMessageBus()
	finch := make(chan error)

	go Run(ADD{userId: "me", amount: 2.0}, mb, ch)
	go Run(BUY{userId: "me", stock: "ABC", cost: 35.1, amount: 1.0}, mb, ch)
	go func() {
		Run(&COMMIT_BUY{userId: "me"}, mb, ch)
		finch <- nil
	}()
	sellT := <-ch
	if sellT.Command != notifyADD && sellT.User_id != "you" && sellT.User_id != "me" {
		t.Fatalf("This transaction should have been an COMMIT_SELL %v", sellT.Command)
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
