package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"log"
)

func main() {
	parent, dispose := goactor.NewParentActor(nil)
	defer dispose(parent)

	abnormalPID := goactor.Spawn(abnormalActor, nil)
	_ = parent.Monitor(abnormalPID)

	err := goactor.Send(abnormalPID, "Hi, I think you should panic :)")
	if err != nil {
		log.Fatal(err)
	}

	parent.Receive(func(msg interface{}) (loop bool) {
		fmt.Println("[!] parent actor received:", msg)
		return false
	})
}

func abnormalActor(actor *goactor.Actor) {
	actor.Receive(func(msg interface{}) (loop bool) {
		fmt.Println("[!] abnormalActor received:", msg)
		panic(msg)
	})
}
