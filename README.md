# Goactor
Golang Actor Model with an Erlang/Elixir API style.

Inspired by Erlang/Elixir's concurrency model, goactor provides a toolkit to spawn and supervise isolated processes that can safely run codes which could panic, without affecting either other processes or the rest of your program.

#### Not familiar with Actor Model?
Well, simply put based on definition, the Actor Model defines a concurrency model where an Actor is a process completely isolated from the rest of the program, with no memory being shared between the actors, and the only way to communicate with them is through exchanging messages.

## Todo:
* Complete this README
* Distributed actors (using gRPC?)
* Finish the TODOs in the code 
* Writing tests for the supervisor package
* Refactoring (simplify) the supervisor package 
* Refactoring error messages and comments, also comment out the remaining parts.

## How to install it?
 Using `go get` command in your terminal: `go get -u github.com/hedisam/goactor` or if your project has go modules enabled, just import the package `github.com/hedisam/goactor` and then run `go mod tidy`.

## How to use it?
To spawn a process, we need a function which we'd like to call it Actor function since it needs to have the signature of `goactor.ActorFunc` that looks like this:
```golang
type ActorFunc func(actor *Actor)
```
So any function that implements `goactor.ActorFunc` type will fit our needs.

Let's spawn an echo actor and send it a message:

```golang
package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"log"
	"time"
)

func main() {
	// Spawn returns a PID which can be used to send message to the actor
	pid := goactor.Spawn(echo, nil) // [*1]

	// Send takes the pid of the target actor and a message to send
	err := goactor.Send(pid, "Hello, Actor") // [*2]
	if err != nil {
		log.Println(err)
	}

	// let's wait a bit to make sure echo gets the chance of printing the message
	<-time.After(100 * time.Millisecond) // [*3]
}

// echo is our actor function (aka our actor's body)
func echo(actor *goactor.Actor) {
	// this is the body of our actor. echo actor will be running until this function returns.
	
	// [*4]
	// Receive accepts a MessageHandler function. It will block until its message handler
	// returns false.
	err := actor.Receive(func(message interface{}) (loop bool) {
		fmt.Printf("[+] Echo received: %v\n", message)
		return false
	})
	if err != nil {
		log.Println(err)
	}
}
```
So given the actor function, echo, we use the Spawn function to start an isolated actor. The spawned actor is just the echo function, and it will be alive and running as long as the echo function has not returned.

To read more on this example, head to the wiki section page [Basics](https://github.com/hedisam/goactor/wiki/Basics#the-same-first-example-from-the-readme-but-with-more-details) where you can find in deep explanation on the numbered lines of the code (e.g. `[*1]`).
