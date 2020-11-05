package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"time"
)

func main() {
	parentActorLinked()
}

func parentActorLinked() {
	parent, dispose := goactor.NewParentActor(nil)
	defer dispose(parent)

	pid := goactor.Spawn(firstActor, nil)
	_ = parent.Link(pid)

	_ = goactor.Send(pid, "panic")

	parent.Receive(func(msg interface{}) (loop bool) {
		fmt.Println("[!] parent received:", msg)
		return false
	})
}

func multipleLinkedActors() {
	firstPID := goactor.Spawn(firstActor, nil)
	secondPID := goactor.Spawn(secondActor, nil)
	thirdPID := goactor.Spawn(secondActor, nil)
	fourthPID := goactor.Spawn(secondActor, nil)

	// asking the second actor to get linked to the first one
	_ = goactor.Send(secondPID, firstPID)
	_ = goactor.Send(thirdPID, firstPID)
	_ = goactor.Send(fourthPID, firstPID)

	// send a message to the panicActor to make it panic
	_ = goactor.Send(firstPID, "A random message to make you panic!")

	time.Sleep(3 * time.Second)
}

func secondActor(actor *goactor.Actor) {
	actor.Receive(func(msg interface{}) (loop bool) {
		pid, _ := msg.(*goactor.PID)
		_ = actor.Link(pid)
		return true
	})
}

func firstActor(actor *goactor.Actor) {
	actor.Receive(func(msg interface{}) bool {
		panic(msg)
	})
}
