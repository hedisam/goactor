package mailbox

import (
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
