package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"log"
	"time"
)

func main() {
	echoPID := goactor.Spawn(echo, nil)

	_ = goactor.Send(echoPID, "it's a direct message")

	err := goactor.SendNamed("echo", "it's a msg to an actor named echo which does not exist")
	if err != nil {
		log.Println(err)
	}

	goactor.Register("echo", echoPID)

	err = goactor.SendNamed("echo", "a msg to a named actor named echo")
	if err != nil {
		log.Println(err)
	}

	_ = goactor.SendNamed("echo", "shutdown")

	time.Sleep(1 * time.Second)
}

func echo(actor *goactor.Actor) {
	actor.Receive(func(msg interface{}) (loop bool) {
		if msg == "shutdown" {
			return false
		}
		fmt.Println("echo actor:", msg)
		return true
	})
}
