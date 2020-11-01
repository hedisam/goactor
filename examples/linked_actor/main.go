package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"time"
)

func main() {
	parent, dispose := goactor.NewParentActor(nil)
	defer dispose()

	firstPID := goactor.Spawn(firstActor, nil)
	//secondPID := goactor.Spawn(secondActor, nil)

	// linking parent actor to the first one
	parent.Link(firstPID)

	// asking the second actor to get linked to the first one
	//_ = goactor.Send(secondPID, firstPID)

	// send a message to the panicActor to make it panic
	_ = goactor.Send(firstPID, "A random message to make you panic!")

	parent.Receive(func(msg interface{}) (loop bool) {
		fmt.Printf("parent actor received: %v\n", msg)
		return false
	})

	time.Sleep(3 * time.Second)
}

func secondActor(actor *goactor.Actor) {
	//actor.SetTrapExit(true)
	actor.Receive(func(msg interface{}) (loop bool) {
		switch message := msg.(type) {
		case goactor.PID:
			actor.Link(message)
			return true
		default:
			fmt.Printf("second actor %v received a message: %v\n", actor.ID(), msg)
			return false
		}
	})
}

func firstActor(actor *goactor.Actor) {
	actor.Receive(func(msg interface{}) bool {
		panic(msg)
	})
}
