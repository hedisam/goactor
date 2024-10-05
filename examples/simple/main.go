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
		timeout: 4 * time.Second,
	}
	_, _ = goactor.Spawn(ctx, simpleActor)

	err := goactor.Send(ctx, simpleActor, "Hey what's up?")
	if err != nil {
		panic(err)
	}
	time.Sleep(1200 * time.Millisecond)
	err = goactor.Send(ctx, simpleActor.self, "Here's my second Hi :)")
	if err != nil {
		panic(err)
	}

	err = goactor.Register(":simple", simpleActor)
	if err != nil {
		panic(err)
	}

	err = goactor.Send(ctx, goactor.NamedPID(":simple"), "You are now registered :yay")
	if err != nil {
		panic(err)
	}

	err = goactor.Send(ctx, goactor.NamedPID(":not_found"), "This message won't make it")
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

func (s *SimpleActor) AfterFunc() (time.Duration, goactor.AfterFunc) {
	return s.timeout, func(ctx context.Context) error {
		fmt.Printf("[!] SimpleActor: no message received after %s; %s elapsed since first msg\n", s.timeout, time.Since(*s.start))
		s.wg.Done()
		return nil
	}
}

func (s *SimpleActor) Init(ctx context.Context, pid *goactor.PID) error {
	s.self = pid
	fmt.Printf("[!] Self PID set: %s\n", pid)
	return nil
}

func (s *SimpleActor) PID() *goactor.PID {
	return s.self
}
