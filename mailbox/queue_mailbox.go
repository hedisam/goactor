package mailbox

import (
	"github.com/Workiva/go-datastructures/queue"
	"runtime"
	"sync/atomic"
	"time"
)

type queueMailbox struct {
	userMsgQueue        *queue.RingBuffer
	sysMsgQueue         *queue.RingBuffer
	sendTimeout         time.Duration
	goSchedulerInterval uint16
	disposed 			uint32
}

func NewQueueMailbox(userMailboxCap, sysMailboxCap int, sendTimeout time.Duration, schedulerInterval uint16) *queueMailbox {
	return &queueMailbox{
		userMsgQueue:        queue.NewRingBuffer(uint64(userMailboxCap)),
		sysMsgQueue:         queue.NewRingBuffer(uint64(sysMailboxCap)),
		sendTimeout:         sendTimeout,
		goSchedulerInterval: schedulerInterval,
	}
}

func (m *queueMailbox) Receive(msgHandler, sysMsgHandler func(interface{}) bool) error {
	return m.ReceiveWithTimeout(0, msgHandler, sysMsgHandler)
}

func (m *queueMailbox) ReceiveWithTimeout(timeout time.Duration, msgHandler, sysMsgHandler func(interface{}) bool) error {
	var i uint16
	var start time.Time
	if timeout > 0 {
		start = time.Now()
	}
	for {
		if atomic.LoadUint32(&m.disposed) == 1 {
			return ErrMailboxClosed
		}

		// our first priority is system messages
		if m.sysMsgQueue.Len() > 0 {
			sysMsg, err := m.sysMsgQueue.Get()
			if err != nil {
				return ErrMailboxClosed
			}
			if !sysMsgHandler(sysMsg) {
				// stop looping through the mailbox
				return nil
			}
			if timeout > 0 {
				// we just processed a message so reset the start time
				start = time.Now()
			}
		}

		// checking user mailbox
		if m.userMsgQueue.Len() > 0 {
			msg, err := m.userMsgQueue.Get()
			if err != nil {
				return ErrMailboxClosed
			}
			if !msgHandler(msg) {
				// stop looping through the mailbox
				return nil
			}

			if timeout > 0 {
				// we just processed a message so reset the start time
				start = time.Now()
			}
		}

		if timeout > 0 && time.Since(start) >= timeout {
			msgHandler(TimedOut{})
			return nil
		}

		// allowing other goroutines to run
		if m.goSchedulerInterval > 0 {
			if i%m.goSchedulerInterval == 0 {
				runtime.Gosched()
				// reset the counter
				i = 1
				continue
			}
			i++
		}
	}
}

func (m *queueMailbox) PushMessage(msg interface{}) error {
	return m.push(m.userMsgQueue, msg)
}

func (m *queueMailbox) PushSystemMessage(msg interface{}) error {
	return m.push(m.sysMsgQueue, msg)
}

func (m *queueMailbox) push(queue *queue.RingBuffer, msg interface{}) error {
	var start time.Time
	if m.sendTimeout > 0 {
		start = time.Now()
	}
	for {
		ok, err := queue.Offer(msg)
		if err != nil {
			return ErrMailboxClosed
		}
		if ok {
			return nil
		}
		if m.sendTimeout > 0 && time.Since(start) >= m.sendTimeout {
			return ErrMailboxTimeout
		}
		runtime.Gosched()
	}
}

func (m *queueMailbox) Dispose() {
	atomic.StoreUint32(&m.disposed, 1)
	m.sysMsgQueue.Dispose()
	m.userMsgQueue.Dispose()
}
