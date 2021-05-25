# GoActor
Go Actor Model with an Erlang/Elixir API style.

Inspired by Erlang/Elixir's concurrency model, goactor provides a toolkit to spawn and supervise isolated processes that can safely run codes which could panic, without affecting either other processes or the rest of your program.

#### Not familiar with Actor Model?
Well, simply put based on definition, the Actor Model defines a concurrency model where an Actor is a process completely isolated from the rest of the program, with no memory being shared between the actors, and the only way to communicate with them is through exchanging messages.

## Contents
* [Todo](https://github.com/hedisam/goactor#todo)
* [How to install it?](https://github.com/hedisam/goactor#how-to-install-it)
* [How to use it?](https://github.com/hedisam/goactor#how-to-use-it)
* [A basic example](https://github.com/hedisam/goactor#a-basic-example)
* [Monitoring & Parent actor](https://github.com/hedisam/goactor#monitoring--parent-actor)
* [Link to another actor](https://github.com/hedisam/goactor#link-to-another-actor)
	* [Trap Exit functionality](https://github.com/hedisam/goactor#link--trap-exit)
* [Register an actor with a name](https://github.com/hedisam/goactor#register-an-actor-with-a-name)
* [Supervisors & Supervision tree](https://github.com/hedisam/goactor/blob/master/README.md#supervisor--supervision-tree)

## Todo:
* Complete this README
* Distributed actors (using gRPC?)
* Finish the TODOs in the code 
* Writing tests for the supervisor package
* Refactoring (simplify) the supervisor package 
* Refactoring error messages and comments, also comment out the remaining parts.
* Document the project
* Logging 

## How to install it?
 Using `go get` command in your terminal: `go get -u github.com/hedisam/goactor` or if your project has go modules enabled, just import the package `github.com/hedisam/goactor` and then run `go mod tidy`.

## How to use it?
To spawn a process, we need a function which we'd like to call it Actor function since it needs to have the signature of `goactor.ActorFunc` that looks like this:
```golang
type ActorFunc func(actor *Actor)
```
So any function that implements `goactor.ActorFunc` type will fit our needs.

### A basic example
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

_[*1]_ `goactor.Spawn` takes two parameters, the first is our actor func, and the second one is a `mailbox builder` which we will talk about it later. It returns a `PID` (process identifier) which is used to send messages to the actor.

_[*2]_ `goactor.Send` takes two parameters as well, the target actor's PID and the message to send. It returns an error in case of a disposed mailbox (which happens when the actor is no longer alive), or due to a mailbox's push timeout which you can set using a mailbox builder when you spawn your actor.

_[*3]_ As you know, the **main goroutine** doesn't wait for other concurrent goroutines to finish, therefore you may not see the result of their work. Since **the building block of our actors are goroutines**, we need to make sure that our echo actor has enough time to show the result of its work.

_[*4]_ `goactor.Receive` accepts a message handler function with a signature of `goactor.MessageHandler`. Here's its signature: `func(message interface{}) (loop bool)`. The method `goactor.Receive` blocks the routine and processes the messages saved in the mailbox by dispatching them to our `MessageHandler` one by one until the handler returns `false`. If no messages are in the mailbox, it will listen for new ones to be received and processed. Your `MessageHandler` should return `true` if you want to process more messages. Note that returning `false` inside your `MessageHandler` does not stop the mailbox from receiving new messages. `goactor.Receive` returns an error if the mailbox is closed.

**NOTE:** no goroutines should be spawned inside an actor (aka actor function) because an unhandled panic in that goroutine will spread out to the entire program.

### Monitoring & Parent actor
To receive a message you need to use an actor's receive method (e.g. `actor.Receive(...)` and that implies you to be within the actor's `ActorFunc` body but sometimes you want to receive and process a message outside of an actor's boundary. That's when you can embrace a `Parent` actor which doesn't need to be spawned.

Here we create a parent actor to `Monitor` another one that is supposed to panic. So we get notified when it panics/exits.

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
	// NewParentActor accepts a mailbox builder function (MailboxBuilderFunc) as its parameter, just like the
	// second argument of goactor.Spawn
	// MailboxBuilderFunc can be used to provide a customized mailbox to the actor.
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
		// after panic-ing, iWillPanic actor broadcasts a specific system message of type sysmsg.SystemMessage so
		// any actor that's linked or monitoring this one will receive the message.
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

### Link to another actor
We use the previous example but instead of monitoring the 'iWillPanic' actor, we `Link` our parent actor to it. 
The difference between `Monitor` and `Link` is that if an actor exit abnormally (e.g. panics), all of its linked actors will exit, too, while a monitor actor only gets notified with no harm in such situations. Also, `Link` creates a two-way relationship between the actors, so either one if exits abnormally causes the other to exit, too.

Here we expect our parent actor to exit along with its linked 'iWillPanic' one:

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
	defer dispose()

	iPanicPID := goactor.Spawn(iWillPanic, nil)

	///////////////////////////
	// the following line of code has changed compared to the previous sample code
	//////////////////////////
	// our parent actor will panic/exit if its linked actor panics
	err := parent.Link(iPanicPID)
	if err != nil {
		log.Println(err)
		return
	}

	// let's send a message to iWillPanic actor which is supposed to panic as soon as it wants to process the message
	err = goactor.Send(iPanicPID, "No matter what message it is, it cause you to panic")
	if err != nil {
		log.Println(err)
		return
	}
	
	/////////////////
	// the following receive method should not get any messages since the parent actor will exit along
	// with 'iWillPanic'
	////////////////
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
		// after panic-ing, iWillPanic actor broadcasts a specific system message of type sysmsg.SystemMessage so
		// any actor that's linked or monitoring this one will receive the message.
		panic(message)
	})
}

```
And here's the output:
```
2021/05/05 18:10:20 dispose: actor 48804599-d184-40a7-85fc-3973f0e3f729 had a panic, reason: No matter what message it is, it cause you to panic
2021/05/05 18:10:20 actor 6173556d-8ae4-4e6c-a35e-f0b51e0ed8e9 received an abnormal exit message from 48804599-d184-40a7-85fc-3973f0e3f729, reason: No matter what message it is, it cause you to panic
```
The actor with id `48804599-d184-40a7-85fc-3973f0e3f729` is the 'iWillPanic' actor whose panic has been handled.
In the second log message you can see the actor with id `6173556d-8ae4-4e6c-a35e-f0b51e0ed8e9` which is our parent actor. The log message shows that it has exited because of receiving an abnormal exit message that is due to being linked to an actor that has panic-ed.

Note that the second log message has been printed by the parent actor's internal methods and not by its `ReceiveWithTimeout` written in the sample code.

#### Link & Trap Exit
When an actor exits or gets terminated, it notifies its linked (and monitor) actors by broadcasting a system message which will be handled internally by the actors that receive it. But what if you wanted to handle this kind of system messages by yourself? Well, you can just do that by trapping exit messages.<br/> 
From the example below, you can see that the parent actor will not panic as it did in the previous example since we've set trap exit to true for it.

```golang
package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/mailbox"
	"github.com/hedisam/goactor/sysmsg"
	"log"
	"time"
)

func main() {
	// let's provide each of our actors a different customized mailbox
	queueMailboxBuilder := func() goactor.Mailbox {
		return mailbox.NewQueueMailbox(5, 5, 100 * time.Millisecond, mailbox.DefaultGoSchedulerInterval)
	}
	chanMailboxBuilder := func() goactor.Mailbox {
		return mailbox.NewChanMailbox(5, 5, 100 * time.Millisecond)
	}

	parent, dispose := goactor.NewParentActor(queueMailboxBuilder)
	defer dispose()

	// by trapping exit messages the parent actor can survive if its linked actors panic or exit abnormally
	parent.SetTrapExit(true)

	iPanicPID := goactor.Spawn(iWillPanic, chanMailboxBuilder)

	err := parent.Link(iPanicPID)
	if err != nil {
		log.Println(err)
		return
	}

	// let's send a message to iWillPanic actor which is supposed to panic as soon as it wants to process the message
	err = goactor.Send(iPanicPID, "No matter what message it is, it cause you to panic")
	if err != nil {
		log.Println(err)
		return
	}

	// we expect to receive a system message since the parent actor is trapping exit messages
	err = parent.ReceiveWithTimeout(time.Millisecond * 100, func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case sysmsg.SystemMessage:
			fmt.Printf("[+] parent received a system message from %s: %v\n", msg.Sender().ID(), msg)
		}
		return false
	})
	// receive timeout is not going to get triggered so this err should be nil
	if err != nil {
		log.Println(err)
	}

	fmt.Println("[!] parent actor is ok")
}

func iWillPanic(actor *goactor.Actor) {
	_ = actor.Receive(func(message interface{}) (loop bool) {
		// after panic-ing, iWillPanic actor broadcasts a specific system message of type sysmsg.SystemMessage so any
		// actor that's linked or monitoring this one will receive the message.
		panic(message)
	})
}

```
Output:
```
[+] parent received a system message from caba2187-b9bd-4cab-9ce6-e6ebe3bce584: {0xc000040180 No matter what message it is, it cause you to panic <nil>}
[!] parent actor is ok
2021/05/11 18:18:34 dispose: actor caba2187-b9bd-4cab-9ce6-e6ebe3bce584 had a panic, reason: No matter what message it is, it cause you to panic
```
The first two lines are `fmt` messages printed by the `parent` actor and the last one is a `log` message which belongs to the `iWillPanic` actor that shows it has panic-ed. The `log` message should've been printed in the first line but be aware that printing `log` messages take a bit longer compared to normal `fmt` ones so don't get confused by the order of the print.<br/>
Nevertheless, you can see that the `parent` actor has not panic-ed and has exited normally.
### Register an actor with a name
Its How-to-do to be added in the next following days
### Supervisor & Supervision tree
Its How-to-do to be added in the next following days
