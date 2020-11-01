package main

import (
	"go/types"
	"log"
)

type handler func(msg interface{})

func main() {
	err := recover()

	switch r := err.(type) {
	case types.Nil:
		log.Println("type is nil")
	default:
		log.Println("type:", r)
	}
}

func create() (func(handler), func()) {
	return receive, dispose
}

func receive(h handler) {
	h("hi")
}

func dispose() {
	if r := recover(); r != nil {
		log.Println("got a panic:", r)
	}
}