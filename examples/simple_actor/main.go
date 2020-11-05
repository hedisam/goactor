package main

import (
	"fmt"
	"github.com/hedisam/goactor"
)

type ShutdownMsg struct {
	sender *goactor.PID
}
type PingMsg struct {
	text   string
	sender *goactor.PID
}

func main() {
	parent, dispose := goactor.NewParentActor(nil)
	defer dispose(parent)

	pingMsg := PingMsg{text: "ping", sender: parent.Self()}

	pid := goactor.Spawn(ping, nil)
	_ = goactor.Send(pid, pingMsg)
	_ = goactor.Send(pid, pingMsg)
	_ = goactor.Send(pid, ShutdownMsg{parent.Self()})

	parent.Receive(func(msg interface{}) (loop bool) {
		fmt.Println("parent:", msg)
		if _, ok := msg.(ShutdownMsg); ok {
			return false
		}
		return true
	})
}

func ping(actor *goactor.Actor) {
	actor.Receive(func(msg interface{}) (loop bool) {
		switch m := msg.(type) {
		case ShutdownMsg:
			_ = goactor.Send(m.sender, ShutdownMsg{actor.Self()})
			return false
		case PingMsg:
			if err := goactor.Send(m.sender, "pong"); err != nil {
				panic(err)
			}
			return true
		default:
			fmt.Println("ping, unknown msg:", msg)
			return true
		}
	})
}

func echo(actor *goactor.Actor) {
	actor.Receive(func(msg interface{}) (loop bool) {
		if msg == "shutdown" {
			fmt.Println("echo: shutting down")
			return false
		}
		fmt.Println("echo:", msg)
		return true
	})
}
