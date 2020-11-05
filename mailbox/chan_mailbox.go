package mailbox

import (
	"time"
)

type chanMailbox struct {
	userMsgChan chan interface{}
	sysMsgChan  chan interface{}

	done chan struct{}
}

func NewChanMailbox() *chanMailbox {
	return &chanMailbox{
		userMsgChan: make(chan interface{}, DefaultUserMailboxCap),
		sysMsgChan:  make(chan interface{}, DefaultSysMailboxCap),
		done:        make(chan struct{}),
	}
}

func (m *chanMailbox) Receive(msgHandler, sysMsgHandler func(interface{}) bool) {
	for {
		select {
		case sysMsg := <-m.sysMsgChan:
			if !sysMsgHandler(sysMsg) {
				// terminate if we should not continue looping through the mailbox
				return
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

func (m *chanMailbox) PushMessage(msg interface{}) error {
	select {
	case <-m.done:
		return ErrMailboxClosed
	case m.userMsgChan <- msg:
		return nil
	case <-time.After(DefaultMailboxTimeout):
		return ErrMailboxTimeout
	}
}

func (m *chanMailbox) PushSystemMessage(msg interface{}) error {
	select {
	case <-m.done:
		return ErrMailboxClosed
	case m.sysMsgChan <- msg:
		return nil
	case <-time.After(DefaultMailboxTimeout):
		return ErrMailboxTimeout
	}
}

func (m *chanMailbox) Dispose() {
	select {
	case <-m.done:
		// it's already closed.
		return
	default:
		close(m.done)
	}
}
