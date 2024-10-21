package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/examples/require"
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
	pid, err := goactor.Spawn(ctx, simpleActor)
	require.NoError(err)

	err = goactor.Send(ctx, pid, "Hey what's up?")
	require.NoError(err)

	time.Sleep(1200 * time.Millisecond)
	err = goactor.Send(ctx, simpleActor, "Here's my second Hi :)")
	require.NoError(err)

	err = goactor.Register(":simple", pid)
	require.NoError(err)

	err = goactor.Send(ctx, goactor.Named(":simple"), "You are now registered :yay")
	require.NoError(err)

	err = goactor.Send(ctx, goactor.Named(":404"), "This message won't make it")
	require.Error(err, "expected to get an error when sending to a non existent named actor")
	fmt.Printf("[!] SendNamed Error: %s\n", err)

	wg.Wait()
}

type SimpleActor struct {
	self    *goactor.PID
	wg      *sync.WaitGroup
	timeout time.Duration
	start   *time.Time
}

func (s *SimpleActor) Receive(_ context.Context, msg any) error {
	if s.start == nil {
		now := time.Now()
		s.start = &now
	}
	fmt.Printf("[+] SimpleActor: %+v %s elapsed from start\n", msg, time.Since(*s.start))
	return nil
}

func (s *SimpleActor) AfterFunc() (time.Duration, goactor.AfterFunc) {
	return s.timeout, func(ctx context.Context) error {
		fmt.Printf("[!] SimpleActor: no message received after %s; %s elapsed since first msg\n", s.timeout, time.Since(*s.start))
		s.wg.Done()
		return nil
	}
}

func (s *SimpleActor) Init(ctx context.Context) error {
	s.self = goactor.Self()
	fmt.Printf("[!] Self PID set: %s\n", s.self.Ref())
	return nil
}

func (s *SimpleActor) PID() *goactor.PID {
	return s.self
}
