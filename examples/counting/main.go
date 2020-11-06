package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

const unit = 50 * 10000
const count = 10 * 1000000

var counterPID *goactor.PID

func main() {
	aggregatedCounting()
	//singleActorCounting()
}

func aggregatedCounting() {
	counterPID = goactor.Spawn(counter, nil)

	n := count / unit
	start := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(n)

	process := func() {
		pid := goactor.Spawn(delegator, nil)
		wg.Done()
		select {
		case <-start:
		}
		for i := 0; i < unit; i++ {
			if err := goactor.Send(pid, "count me"); err != nil {
				log.Println("process:", err)
			}
		}
	}

	fmt.Printf("[!] starting %d goroutines & actors\n", n)
	for i := 0; i < n; i++ {
		go process()
	}

	wg.Wait()
	fmt.Println("[!] start broadcasting...")
	close(start)

	wait()
}

func delegator(actor *goactor.Actor) {
	i := 0
	actor.Receive(func(msg interface{}) (loop bool) {
		i++
		if err := goactor.Send(counterPID, msg); err != nil {
			log.Println("delegator:", err)
		}
		if i == unit {
			return false
		}
		return true
	})
}

func singleActorCounting() {
	pid := goactor.Spawn(counter, goactor.DefaultQueueMailbox)

	n := count / unit
	start := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(n)

	process := func() {
		wg.Done()
		select {
		case <-start:
		}
		for i := 0; i < unit; i++ {
			if err := goactor.Send(pid, "count me"); err != nil {
				log.Println(err)
			}
		}
	}

	fmt.Printf("[!] starting %d goroutines\n", n)
	for i := 0; i < n; i++ {
		go process()
	}

	wg.Wait()
	fmt.Println("[!] start broadcasting...")
	close(start)

	wait()
}

func counter(actor *goactor.Actor) {
	i := 0
	now := time.Now()
	actor.Receive(func(msg interface{}) (loop bool) {
		i++
		if i == count {
			elapsed := time.Since(now)
			fmt.Printf("[+] received %d messages in %v\n", i, elapsed)
			return false
		}
		if i%unit == 0 {
			fmt.Println("[-] received:", i)
		}
		return true
	})
}

func wait() {
	signalChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, os.Kill)
	go func() {
		<-signalChan
		fmt.Println("[!] CTRL+C")
		close(done)
	}()
	<-done
}
