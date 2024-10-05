package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hedisam/goactor"
)

func main() {
	fmt.Println("----- Simple Actor ----")
	ctx := context.Background()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	simpleActor := &SimpleActor{
		wg:      wg,
		timeout: 2 * time.Second,
	}
	_ = goactor.Spawn(ctx,
		simpleActor.Receive,
		goactor.WithInitFunc(simpleActor.Init),
		goactor.WithAfterFunc(2*simpleActor.timeout, simpleActor.After),
	)

	err := goactor.Send(ctx, simpleActor.self, "Hey what's up?")
	if err != nil {
		panic(err)
	}
	time.Sleep(1200 * time.Millisecond)
	err = goactor.Send(ctx, simpleActor, "Here's my second Hi :)")
	if err != nil {
		panic(err)
	}

	err = goactor.Register(":simple", simpleActor)
	if err != nil {
		panic(err)
	}

	err = goactor.SendNamed(ctx, ":simple", "You are now registered :yay")
	if err != nil {
		panic(err)
	}

	err = goactor.SendNamed(ctx, ":not_found", "This message won't make it")
	if err == nil {
		log.Fatal("Expected to get error when sending to a non existent named actor but got nil")
	}
	fmt.Printf("[!] SendNamed Error: %s\n", err)

	wg.Wait()
}

type SimpleActor struct {
	self    *goactor.PID
	wg      *sync.WaitGroup
	timeout time.Duration
	start   *time.Time
}

func (s *SimpleActor) Receive(_ context.Context, msg any) (loop bool, err error) {
	if s.start == nil {
		now := time.Now()
		s.start = &now
	}
	fmt.Printf("[+] SimpleActor: %+v %s elapsed from start\n", msg, time.Since(*s.start))
	return true, nil
}

func (s *SimpleActor) After(ctx context.Context) error {
	fmt.Printf("[!] SimpleActor: no message received after %s; %s elapsed since first msg\n", s.timeout, time.Since(*s.start))
	s.wg.Done()
	return nil
}

func (s *SimpleActor) Init(ctx context.Context, pid *goactor.PID) {
	s.self = pid
	fmt.Printf("[!] Self PID set: %s\n", pid)
}

func (s *SimpleActor) PID() *goactor.PID {
	return s.self
}
