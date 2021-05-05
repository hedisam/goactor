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

#### Monitoring & Parent actor

```golang
package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/sysmsg"
	"log"
	"time"
)

func main() {
	parent, dispose := goactor.NewParentActor(nil)
	defer dispose() // don't forget to defer the dispose function of a parent actor

	iPanicPID := goactor.Spawn(iWillPanic, nil)

	// when you monitor another actor, you expect to get notified about anything (bad) that happens to the target actor.
	err := parent.Monitor(iPanicPID)
	if err != nil {
		log.Println(err)
		return
	}

	// let's send a message to the iWillPanic actor which is supposed to panic as soon as it wants to process the message
	err = goactor.Send(iPanicPID, "No matter what message it is, it cause you to panic")
	if err != nil {
		log.Println(err)
		return
	}

	// ReceiveWithTimeout returns a timeout error if no messages came through by the specified timeout
	err = parent.ReceiveWithTimeout(time.Millisecond * 100, func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case sysmsg.SystemMessage:
			fmt.Printf("[+] parent received a system message from %s: %v", msg.Sender().ID(), msg)
		}
		return false
	})
	if err != nil {
		log.Println(err)
	}
}

func iWillPanic(actor *goactor.Actor) {
	_ = actor.Receive(func(message interface{}) (loop bool) {
		// after panic-ing, iWillPanic actor broadcasts a specific system message of type sysmsg.SystemMessage so any actor that's linked
		// or monitoring this one will receive the message.
		panic(message)
	})
}

```
And here's the output:
```
2021/05/05 16:32:04 dispose: actor a8895eb4-302c-4ed9-86f5-3aada8b5c8c6 had a panic, reason: No matter what message it is, it cause you to panic
[+] parent received a system message from a8895eb4-302c-4ed9-86f5-3aada8b5c8c6: {0xc00016a140 No matter what message it is, it cause you to panic <nil>}
```
The first line of the output is a log message internally printed by the panic-ed actor, and the second one is the message received and printed by our parent actor.

You don't need to spawn an `ActorFunc` function to create a `parent` actor since it works and runs in the same goroutine you use to create it (in our example it would be the main goroutine). So it's important to know it can only survive from panics that happen within its goroutine's boundaries.
