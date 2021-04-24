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

var ErrMailboxClosed = fmt.Errorf("target mailbox is closed/disposed")
var ErrMailboxEnqueueTimeout = fmt.Errorf("mailbox send timeout")
var ErrMailboxReceiveTimeout = fmt.Errorf("mailbox receive timeout")

type TimedOut struct{}

func (t TimedOut) Error() string {
	return "receive timeout"
}
