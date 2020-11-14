package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/supervisor"
	"log"
	"time"
)

func main() {
	ref, err := supervisor.Start(
		supervisor.OneForOneStrategyOption(),
		supervisor.NewWorkerSpec("the panicer", supervisor.RestartTransient, toPanic),
	)
	if err != nil {
		log.Fatal(err)
	}

	info, err := ref.ChildrenCount(2 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info)

	for {
		err = goactor.SendNamed("the panicer", "hey you wanna panic?")
		if err != nil {
			log.Println(err)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	info, err = ref.ChildrenCount(2 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info)
}

func toPanic(actor *goactor.Actor) {
	fmt.Println("[!] starting toPanic actor...")
	actor.Receive(func(message interface{}) (loop bool) {
		fmt.Println("[!] toPanic actor received a message -> so I panic")
		panic(message)
	})
}
