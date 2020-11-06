package mailbox

import (
	"fmt"
	"time"
)

const (
	DefaultUserMailboxCap      = 100
	DefaultSysMailboxCap       = 20
	DefaultMailboxTimeout      = 2 * time.Second
	DefaultGoSchedulerInterval = 1000
)

const (
	mailboxProcessing int32 = iota
	mailboxIdle
)

var ErrMailboxClosed = fmt.Errorf("target mailbox is closed")
var ErrMailboxTimeout = fmt.Errorf("mailbox sendTimeout")

type TimedOut struct{}

func (t TimedOut) Error() string {
	return "receive timeout"
}
