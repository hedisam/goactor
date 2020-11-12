package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/mailbox"
	p "github.com/hedisam/goactor/pid"
	"time"
)

var sender *p.PID

func main() {
	future := goactor.NewFutureActor()
	sender = future.Self()

	pid := goactor.Spawn(myPrint, nil)

	_ = goactor.Send(pid, "message #1")

	// just to block
	future.Receive(func(msg interface{}) (loop bool) {
		return false
	})
}

func myPrint(actor *goactor.Actor) {
	actor.ReceiveWithTimeout(2*time.Second, func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case mailbox.TimedOut:
			fmt.Println("[!] myPrint: timeout triggered")
			_ = goactor.Send(sender, "exit")
			return false // it doesn't matter even if you return true. ReceiveWithTimeout exit after a timeout
		default:
			fmt.Println("[+] myPrint:", msg)
			return true
		}
	})
}
