package mailbox

import (
	"github.com/Workiva/go-datastructures/queue"
	"log"
	"runtime"
	"time"
)

type queueMailbox struct {
	userMsgQueue        *queue.RingBuffer
	sysMsgQueue         *queue.RingBuffer
	sendTimeout         time.Duration
	goSchedulerInterval uint16
}

func NewQueueMailbox(capacity int, sendTimeout time.Duration, schedulerInterval uint16) *queueMailbox {
	return &queueMailbox{
		userMsgQueue:        queue.NewRingBuffer(uint64(capacity)),
		sysMsgQueue:         queue.NewRingBuffer(DefaultSysMailboxCap),
		sendTimeout:         sendTimeout,
		goSchedulerInterval: schedulerInterval,
	}
}

func (m *queueMailbox) Receive(msgHandler, sysMsgHandler func(interface{}) bool) {
	var i uint16
	for {
		// our first priority is system messages
		if m.sysMsgQueue.Len() > 0 {
			sysMsg, err := m.sysMsgQueue.Get()
			if err != nil {
				// mailbox is disposed.
				log.Println("receive: mailbox has been disposed, %w", err)
				return
			}
			if sysMsgHandler(sysMsg) {
				// pass the msg to the user
				if !msgHandler(sysMsg) {
					return
				}
			}
		}

		// checking user mailbox
		if m.userMsgQueue.Len() > 0 {
			msg, err := m.userMsgQueue.Get()
			if err != nil {
				// mailbox is disposed.
				log.Println("receive: mailbox has been disposed, %w", err)
				return
			}
			if !msgHandler(msg) {
				return
			}
		}
		// allowing other goroutines to run
		if m.goSchedulerInterval > 0 {
			if i%m.goSchedulerInterval == 0 {
				runtime.Gosched()
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
	m.sysMsgQueue.Dispose()
	m.userMsgQueue.Dispose()
}
