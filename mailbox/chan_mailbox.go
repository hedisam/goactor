package mailbox

import (
	"time"
)

type chanMailbox struct {
	userMsgChan chan interface{}
	sysMsgChan  chan interface{}
	sendTimeout time.Duration
	done        chan struct{}
}

func NewChanMailbox(userMailboxCap, sysMailboxCap int, sendTimeout time.Duration) *chanMailbox {
	return &chanMailbox{
		userMsgChan: make(chan interface{}, userMailboxCap),
		sysMsgChan:  make(chan interface{}, sysMailboxCap),
		done:        make(chan struct{}),
		sendTimeout: sendTimeout,
	}
}

func (m *chanMailbox) Receive(msgHandler, sysMsgHandler func(interface{}) bool) {
	// we could've delegate this to m.ReceiveWithTimeout(0, msgHandler, sysMsgHandler),
	// but for the sake of efficiency, we wouldn't.
	for {
		select {
		case sysMsg := <-m.sysMsgChan:
			if sysMsgHandler(sysMsg) {
				// pass the sysMsg to the user
				if !msgHandler(sysMsg) {
					return
				}
			}
		case msg := <-m.userMsgChan:
			if !msgHandler(msg) {
				return
			}
		case <-m.done:
			return
		}
	}
}

func (m *chanMailbox) ReceiveWithTimeout(timeout time.Duration, msgHandler, sysMsgHandler func(interface{}) bool) {
	if timeout <= 0 {
		m.Receive(msgHandler, sysMsgHandler)
		return
	}
	var ticker = time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case sysMsg := <-m.sysMsgChan:
			if sysMsgHandler(sysMsg) {
				// pass the sysMsg to the user
				if !msgHandler(sysMsg) {
					return
				}
			}
		case msg := <-m.userMsgChan:
			if !msgHandler(msg) {
				return
			}
		case <-m.done:
			return
		case <-ticker.C:
			msgHandler(TimedOut{})
			return
		}

		ticker.Reset(timeout)
		// the next tick could've already been triggered. so drain the channel to prevent unwanted ticks.
		drainChan(ticker.C)
	}
}

func (m *chanMailbox) PushMessage(msg interface{}) error {
	return m.push(m.userMsgChan, msg)
}

func (m *chanMailbox) PushSystemMessage(msg interface{}) error {
	return m.push(m.sysMsgChan, msg)
}

func (m *chanMailbox) push(msgChan chan<- interface{}, msg interface{}) error {
	var timer *time.Timer
	timeoutChan := make(<-chan time.Time, 1)

	if m.sendTimeout > 0 {
		timer = time.NewTimer(m.sendTimeout)
		timeoutChan = timer.C
	}

	if timer != nil {
		defer timer.Stop()
	}

	select {
	case <-m.done:
		return ErrMailboxClosed
	case msgChan <- msg:
		return nil
	case <-timeoutChan:
		return ErrMailboxTimeout
	}
}

func (m *chanMailbox) Dispose() {
	select {
	case <-m.done:
		// it's already been closed.
		return
	default:
		close(m.done)
	}
}

func drainChan(timeoutChan <-chan time.Time) {
	select {
	case <-timeoutChan:
	default:
	}
}
