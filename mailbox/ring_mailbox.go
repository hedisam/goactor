package mailbox

import (
	"github.com/Workiva/go-datastructures/queue"
	"runtime"
	"sync/atomic"
	"time"
)

type ringBufferMailbox struct {
	userMsgQueue *queue.RingBuffer
	sysMsgQueue  *queue.RingBuffer
	done         chan struct{}
	signal       chan struct{}
	status       int32
	timeout      time.Duration
}

func NewRingBufferMailbox() *ringBufferMailbox {
	return &ringBufferMailbox{
		userMsgQueue: queue.NewRingBuffer(DefaultUserMailboxCap),
		sysMsgQueue:  queue.NewRingBuffer(DefaultSysMailboxCap),
		done:         make(chan struct{}),
		signal:       make(chan struct{}),
		status:       mailboxIdle,
		timeout:      DefaultMailboxTimeout,
	}
}

func (m *ringBufferMailbox) ReceiveWithTimeout(timeout time.Duration, msgHandler, sysMsgHandler func(interface{}) bool) {
	panic("ring buffer receive with timeout -> implement me")
}

func (m *ringBufferMailbox) Receive(msgHandler, sysMsgHandler func(interface{}) bool) {
	// declare mailbox as idle when we return using the 'return' keyword
	defer m.idle()
	for {
		select {
		case <-m.done:
			// mailbox has been disposed
			return
		case <-m.signal:
			// signal, indicating new messages are in the mailbox
		loop:
			// our first priority is system messages
			for m.sysMsgQueue.Len() > 0 {
				sysMsg, err := m.sysMsgQueue.Get()
				if err != nil {
					// mailbox has been disposed
					return
				}
				if !sysMsgHandler(sysMsg) {
					return
				}
			}
			// check for user messages
			if m.userMsgQueue.Len() > 0 {
				msg, err := m.userMsgQueue.Get()
				if err != nil {
					return
				}
				if !msgHandler(msg) {
					return
				}
			}
			if m.sysMsgQueue.Len() == 0 && m.userMsgQueue.Len() == 0 {
				m.idle()
				continue
			}
			goto loop
		}
	}
}

func (m *ringBufferMailbox) PushMessage(msg interface{}) error {
	return m.push(m.userMsgQueue, msg)
}

func (m *ringBufferMailbox) PushSystemMessage(msg interface{}) error {
	return m.push(m.sysMsgQueue, msg)
}

func (m *ringBufferMailbox) push(queue *queue.RingBuffer, msg interface{}) error {
	var start time.Time
	if m.timeout > 0 {
		start = time.Now()
	}
	// push the message to the queue
	for {
		ok, err := queue.Offer(msg)
		if err != nil {
			return ErrMailboxClosed
		}
		if ok {
			break
		}
		if m.timeout > 0 && time.Since(start) >= m.timeout {
			return ErrMailboxTimeout
		}
		runtime.Gosched()
	}
	// signal the mailbox to process the new message
	if atomic.CompareAndSwapInt32(&m.status, mailboxIdle, mailboxProcessing) {
		select {
		case m.signal <- struct{}{}:
		case <-time.After(m.timeout - time.Since(start)):
			return ErrMailboxTimeout
		case <-m.done:
			return ErrMailboxClosed
		}
	}
	return nil
}

func (m *ringBufferMailbox) idle() {
	atomic.StoreInt32(&m.status, mailboxIdle)
}

func (m *ringBufferMailbox) Dispose() {
	select {
	case <-m.done:
		// already been closed
		return
	default:
		close(m.done)
		m.sysMsgQueue.Dispose()
		m.userMsgQueue.Dispose()
	}
}
