package goactor

import (
	"context"
	"time"
)

type ReceiveFunc = func(ctx context.Context, msg any) (loop bool, err error)
type AfterTimeoutFunc = func(ctx context.Context) error
type InitFunc = func(ctx context.Context, pid *PID)

type actor struct {
	receiveFunc            ReceiveFunc
	initFunc               InitFunc
	afterTimeoutFunc       AfterTimeoutFunc
	receiveTimeoutDuration time.Duration
}

func newActor(fn ReceiveFunc) *actor {
	return &actor{
		receiveFunc:            fn,
		initFunc:               func(ctx context.Context, pid *PID) {},
		afterTimeoutFunc:       func(ctx context.Context) error { return nil },
		receiveTimeoutDuration: 0,
	}
}

type ActorOption func(a *actor)

func WithAfterFunc(d time.Duration, fn AfterTimeoutFunc) ActorOption {
	return func(a *actor) {
		a.receiveTimeoutDuration = d
		a.afterTimeoutFunc = fn
	}
}

func WithInitFunc(fn InitFunc) ActorOption {
	return func(a *actor) {
		a.initFunc = fn
	}
}
