package goactor

import (
	"context"
	"time"
)

// Actor defines the only required method, Receive, for an Actor implementation.
type Actor interface {
	// Receive is called when a message is received.
	// An error can be returned to stop the actor. sysmsg.ReasonNormal can be returned for a normal exit.
	Receive(ctx context.Context, msg any) error
}

// ReceiveFunc is a helper function that can be used for spawning an actor only with the receiver function without
// the need for a struct implementation.
type ReceiveFunc func(ctx context.Context, msg any) error

// Receive implements Actor.
func (r ReceiveFunc) Receive(ctx context.Context, msg any) error {
	return r(ctx, msg)
}

// ActorInitializer defines the optional method, Init, that can be implemented by an Actor which will be called
// upon spawning the Actor to let the user perform any one-time initialisation they need.
type ActorInitializer interface {
	// Init is called before spawning the Actor when the PID is available.
	Init(ctx context.Context) error
}

// AfterFunc defines the After func signature which can be provided to be called if no messages are received within
// a specified time period.
type AfterFunc func(context.Context) error

// ActorAfterFuncProvider defines the optional method, AfterFunc, that can be implemented by an Actor which will be
// called when no messages are received within the specified timeout duration.
type ActorAfterFuncProvider interface {
	// AfterFunc returns a function to be called if no messages are received after the provided timeout duration.
	// The timer is reset each time a message is received.
	AfterFunc() (timeout time.Duration, afterFunc AfterFunc)
}
