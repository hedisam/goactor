package mailbox

import (
	"context"
	"errors"
)

// Define mailbox default values and consts
const (
	DefaultMessagesCap = 100
	DefaultSystemCap   = 20
)

var (
	// ErrClosedMailbox is returned when the mailbox is closed
	ErrClosedMailbox = errors.New("closed mailbox")
	// ErrReceiveTimeout is returned when timeout occurs while listening for incoming messages
	ErrReceiveTimeout = errors.New("receive timeout")
)

type Dispatcher interface {
	PushMessage(ctx context.Context, msg any) error
	PushSystemMessage(ctx context.Context, msg any) error
}

func PushMessage(ctx context.Context, d Dispatcher, msg any) error {
	return d.PushMessage(ctx, msg)
}

func PushSystemMessage(ctx context.Context, d Dispatcher, msg any) error {
	return d.PushSystemMessage(ctx, msg)
}
