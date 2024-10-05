package goactor

import (
	"context"
	"slices"
	"time"
)

type ReceiveFunc = func(ctx context.Context, msg any) (loop bool, err error)
type InitFunc = func(ctx context.Context, pid *PID) error
type ActorOption func(a *AnActor)

// WithAfterFunc configures AnActor with an After func and its timeout.
func WithAfterFunc(d time.Duration, fn AfterFunc) ActorOption {
	return func(a *AnActor) {
		a.receiveTimeoutDuration = d
		a.afterTimeoutFunc = fn
	}
}

// WithInitFunc configures AnActor with an Init function.
func WithInitFunc(fn InitFunc) ActorOption {
	return func(a *AnActor) {
		a.initFunc = fn
	}
}

// AnActor implements the Actor interface to be used when user don't have a custom Actor struct to spawn.
type AnActor struct {
	receiveFunc            ReceiveFunc
	initFunc               InitFunc
	afterTimeoutFunc       AfterFunc
	receiveTimeoutDuration time.Duration
}

// NewActor returns a new instance of AnActor using the provided receiver and config options.
func NewActor(receiver ReceiveFunc, opts ...ActorOption) *AnActor {
	a := &AnActor{
		receiveFunc:      receiver,
		initFunc:         func(ctx context.Context, pid *PID) error { return nil },
		afterTimeoutFunc: func(ctx context.Context) error { return nil },
	}
	for opt := range slices.Values(opts) {
		opt(a)
	}
	return a
}

func (a *AnActor) Receive(ctx context.Context, msg any) (loop bool, err error) {
	return a.receiveFunc(ctx, msg)
}

func (a *AnActor) Init(ctx context.Context, pid *PID) error {
	return a.initFunc(ctx, pid)
}

func (a *AnActor) AfterFunc() (timeout time.Duration, afterFunc AfterFunc) {
	return a.receiveTimeoutDuration, a.afterTimeoutFunc
}
