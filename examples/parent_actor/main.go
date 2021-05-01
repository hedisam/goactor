package main

import (
	"github.com/hedisam/goactor"
	"log"
	"time"
)

func main() {
	parent, dispose := goactor.NewParentActor(nil)
	defer dispose()

	err := goactor.Send(parent.Self(), "I just want you to panic")
	if err != nil {
		log.Fatal(err)
	}

	err = parent.ReceiveWithTimeout(time.Millisecond * 10, func(message interface{}) (loop bool) {
		panic(message)
	})
	if err != nil {
		log.Fatal(err)
	}
}
