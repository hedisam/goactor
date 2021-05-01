package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"log"
	"time"
)

func main() {
	pid := goactor.Spawn(panicee2, nil)
	parent, dispose := goactor.NewParentActor(nil)
	defer dispose()

	err := parent.Monitor(pid)
	if err != nil {
		log.Fatal(err)
	}

	err = parent.ReceiveWithTimeout(time.Millisecond * 10, func(message interface{}) (loop bool) {
		fmt.Println("[!] parent received:", message)
		return false
	})

	if err != nil {
		log.Fatal(err)
	}
	//
	//select {
	//case <-time.After(10 * time.Millisecond):
	//}
}

func panicee2(a *goactor.Actor) {
	panic("panic-ing outside of receive's body")
}

func panicee(a *goactor.Actor) {
	a.Receive(func(message interface{}) (loop bool) {
		fmt.Println("[!] panicee received a message:", message)
		panic(message)
	})
}
