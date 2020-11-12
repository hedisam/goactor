package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	p "github.com/hedisam/goactor/pid"
	"log"
	"time"
)

type Message struct {
	text   string
	sender *p.PID
}

func main() {
	future := goactor.NewFutureActor()
	echoPID := goactor.Spawn(echo, nil)

	err := goactor.Send(echoPID, Message{text: "Hello echo", sender: future.Self()})
	if err != nil {
		log.Fatal(err)
	}
	err = goactor.Send(echoPID, Message{text: "Hello again", sender: future.Self()})

	future.Receive(func(msg interface{}) (loop bool) {
		fmt.Println("[!] future received:", msg)
		return false
	})

	// future actor get disposed when its receive method exit (receive could be called only once
	// this should not be blocking
	future.Receive(func(msg interface{}) (loop bool) {
		fmt.Println("[!] future received - second receive:", msg)
		return false
	})

	time.Sleep(1 * time.Second)
}

func echo(actor *goactor.Actor) {
	actor.Receive(func(msg interface{}) (loop bool) {
		fmt.Println("[+] echo received:", msg)
		switch m := msg.(type) {
		case Message:
			if err := goactor.Send(m.sender, m.text); err != nil {
				log.Println("[!] echo failed to send a reply:", err)
				return false
			}
			return true
		default:
			return false
		}
	})
}
