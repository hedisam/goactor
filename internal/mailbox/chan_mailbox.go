package mailbox

import (
	"context"
	"sync/atomic"
	"time"
)

// ChanMailbox is an actor mailbox that uses go channels under the hood.
type ChanMailbox struct {
	messageCh chan any
	systemCh  chan any
	closedCh  chan struct{}
	closed    atomic.Bool
	t         *time.Timer
}

// NewChanMailbox returns a new instance of ChanMailbox.
func NewChanMailbox() *ChanMailbox {
	return &ChanMailbox{
		messageCh: make(chan any, DefaultMessagesCap),
		systemCh:  make(chan any, DefaultSystemCap),
		closedCh:  make(chan struct{}),
		t:         time.NewTimer(0),
	}
}

// ReceiveTimeout listens for incoming messages to handle them using the provided handler.
// It stops listening for new messages on timeout. The timeout is refreshed each time a new message is received.
func (m *ChanMailbox) ReceiveTimeout(ctx context.Context, d time.Duration) (msg any, sysMsg bool, err error) {
	if d <= 0 {
		return m.Receive(ctx)
	}

	if m.closed.Load() {
		return nil, false, ErrClosedMailbox
	}

	m.resetTimer(d)

	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case <-m.closedCh:
		return nil, false, ErrClosedMailbox
	case msg = <-m.systemCh:
		return msg, true, nil
	case msg = <-m.messageCh:
		return msg, false, nil
	case <-m.t.C:
		return nil, false, ErrReceiveTimeout
	}
}

// Receive listens for incoming messages to handle them using the provided handler.
// It stops listening if the context is canceled and returns ErrMailboxDisposed if the mailbox is disposed.
func (m *ChanMailbox) Receive(ctx context.Context) (msg any, sysMsg bool, err error) {
	if m.closed.Load() {
		return nil, false, ErrClosedMailbox
	}

	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case <-m.closedCh:
		return nil, false, ErrClosedMailbox
	case msg = <-m.systemCh:
		return msg, true, nil
	case msg = <-m.messageCh:
		return msg, false, nil
	}
}

func (m *ChanMailbox) PushMessage(ctx context.Context, msg any) error {
	return m.push(ctx, m.messageCh, msg)
}

func (m *ChanMailbox) PushSystemMessage(ctx context.Context, msg any) error {
	return m.push(ctx, m.systemCh, msg)
}

func (m *ChanMailbox) push(ctx context.Context, msgChan chan<- any, msg any) error {
	if m.closed.Load() {
		return ErrClosedMailbox
	}

	select {
	case <-m.closedCh:
		return ErrClosedMailbox
	case <-ctx.Done():
		return ctx.Err()
	case msgChan <- msg:
		return nil
	}
}

func (m *ChanMailbox) resetTimer(d time.Duration) {
	if !m.t.Stop() {
		<-m.t.C
	}
	m.t.Reset(d)
}

// Close closes the mailbox and stops any message listener.
func (m *ChanMailbox) Close() {
	if m.closed.CompareAndSwap(false, true) {
		close(m.closedCh)
	}
}