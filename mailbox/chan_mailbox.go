package mailbox

import (
	"fmt"
	"time"
)

type chanMailbox struct {
	userMsgChan chan interface{}
	sysMsgChan  chan interface{}

	done chan struct{}
}

func NewChanMailbox() *chanMailbox {
	return &chanMailbox{
		userMsgChan: make(chan interface{}, 100),
		sysMsgChan:  make(chan interface{}, 20),
		done:        make(chan struct{}),
	}
}

func (m *chanMailbox) Receive(msgHandler, sysMsgHandler func(interface{}) bool) {
	for {
		select {
		case sysMsg := <- m.sysMsgChan:
			if !sysMsgHandler(sysMsg) {
				// terminate if we should not continue looping through the mailbox
				return
			}
		case msg := <- m.userMsgChan:
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
		return fmt.Errorf("target mailbox is closed")
	case m.userMsgChan<- msg:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("timeout")
	}
}

func (m *chanMailbox) PushSystemMessage(msg interface{}) error {
	select {
	case <-m.done:
		return fmt.Errorf("target mailbox is closed")
	case m.sysMsgChan<- msg:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("timeout")
	}
}

func (m *chanMailbox) Dispose() {
	close(m.done)
}