package goactor

import (
	"context"
	"time"
)

type ReceiveFunc = func(ctx context.Context, msg any) (loop bool, err error)
type AfterTimeoutFunc = func(ctx context.Context) error
type InitFunc = func(ctx context.Context, pid *PID)

type actorConfig struct {
	receiveFunc            ReceiveFunc
	initFunc               InitFunc
	afterTimeoutFunc       AfterTimeoutFunc
	receiveTimeoutDuration time.Duration
}

func newActorConfig(fn ReceiveFunc) *actorConfig {
	return &actorConfig{
		receiveFunc:            fn,
		initFunc:               func(ctx context.Context, pid *PID) {},
		afterTimeoutFunc:       func(ctx context.Context) error { return nil },
		receiveTimeoutDuration: 0,
	}
}

type ActorOption func(a *actorConfig)

func WithAfterFunc(d time.Duration, fn AfterTimeoutFunc) ActorOption {
	return func(a *actorConfig) {
		a.receiveTimeoutDuration = d
		a.afterTimeoutFunc = fn
	}
}

func WithInitFunc(fn InitFunc) ActorOption {
	return func(a *actorConfig) {
		a.initFunc = fn
	}
}
