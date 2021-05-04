# Goactor
Golang Actor Model with an Erlang/Elixir API style.

Inspired by Erlang/Elixir's concurrency model, goactor provides a toolkit to spawn and supervise isolated processes \
which can safely run codes that could possibly panic, without affecting either other processes nor the rest of your program.

#### Not familiar with Actor model?\
Well, simply put based on definition, Actor model describes a concurrency model where an Actor is a process completely\
isolated from the rest of the program, with no memory being shared between the actors, and the only way to communicate\
with them is through exchanging messages.

## Todo:
* Complete this README
* Distributed actors (using gRPC?)
* Finish the TODOs in the code 
* Writing tests for the supervisor package
* Refactoring (simplify) the supervisor package 
* Refactoring error messages and comments, also comment out the remaining parts.

## How to install it?
 Using `go get` command in your terminal: `go get -u github.com/hedisam/goactor` \
or if your project has go modules enabled, just import the package `github.com/hedisam/goactor` and\
then run `go mod tidy`.

## How to use it?
To spawn a process, we need a function which we'd like to call it an Actor function since it needs to have the\
signature of `goactor.ActorFunc` which looks like this:
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
	// Spawn returns a pid which can be used to send message to the actor
	pid := goactor.Spawn(echo, nil) // [*1]

	// Send takes the pid of the target actor and a message to send
	err := goactor.Send(pid, "Hello, Actor") // [*2]
	if err != nil {
		log.Println(err)
	}

	// let's wait a bit to make sure echo gets the chance of printing the message
	<-time.After(100 * time.Millisecond) // [*3]
}

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
So given the actor function, echo, we use the Spawn function to start an isolated actor. The spawned actor is\
actually just the echo function, and it will be alive and running as long as the echo function has not returned.

_[*1]_ `goactor.Spawn` takes two parameters, the first is our actor func, and the second one is a mailbox builder which\
we will talk about it later. It returns a pid (process identifier) which is used to send message to the actor.

_[*2]_ `goactor.Send` takes two parameters as well, the target actor's pid and the message to send. It returns an error\
in case of a disposed mailbox (which happens when the actor is no longer alive), or due to a mailbox's push timeout which\
you can set using a mailbox builder when you spawn your actor.

_[*3]_ As you might know, the **main goroutine** doesn't wait for other concurrent goroutines to finish, therefore you may not\
see the result of their work. Since **the building block of our actors are goroutines**, we need to make sure that our\
echo actor has enough time to show the result of its work.

_[*4]_ `goactor.Receive` accepts a function with a signature of `goactor.MessageHandler`. Here's its signature:\
`func(message interface{}) (loop bool)`. The method `goactor.Receive` blocks the routine and processes the\
messages saved in the mailbox by dispatching them to our `MessageHandler` one by one until it returns `false`. If no\
messages are in the mailbox, it will listen for new ones to be received and processed. Your `MessageHandler`\
should return `true` if you want to process more messages. Note that returning `false` inside your\
`MessageHandler` does not stop the mailbox from receiving new messages. `goactor.Receive` returns an error if the mailbox has been disposed.

**NOTE:** no goroutines should be spawned inside an actor (aka actor function), because an unhandled panic in that\
goroutine will spread out to the entire of your program.