package mailbox

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestQueueMailbox_Receive(t *testing.T) {
	var length = 5
	m := NewQueueMailbox(length, length, 10 * time.Millisecond, DefaultSysMailboxCap)

	messages := make([]int, length)
	for i := 0; i < length; i++ {
		messages[i] = i+1
	}

	t.Run("user messages receive", func(t *testing.T) {
		for _, msg := range messages {
			err := m.PushMessage(msg)
			assert.Nil(t, err)
		}

		received := make([]int, length)
		i := 0
		err := m.Receive(func(msg interface{}) bool {
			received[i] = msg.(int)
			i++
			if i == length {return false}
			return true
		}, nil)
		if !assert.Nil(t, err) {return}
		assert.EqualValues(t, messages, received)
	})

	t.Run("system messages receive", func(t *testing.T) {
		for _, msg := range messages {
			err := m.PushSystemMessage(msg)
			assert.Nil(t, err)
		}

		received := make([]int, length)
		i := 0
		err := m.Receive(nil, func(msg interface{}) bool {
			received[i] = msg.(int)
			i++
			if i == length {return false}
			return true
		})
		if !assert.Nil(t, err) {return}
		assert.EqualValues(t, messages, received)
	})

	t.Run("receive with timeout", func(t *testing.T) {
		m := NewQueueMailbox(2, 2, 10 * time.Millisecond, DefaultGoSchedulerInterval)

		wg := sync.WaitGroup{}
		wg.Add(1)
		time.AfterFunc(50 * time.Millisecond, func() {
			defer wg.Done()
			err := m.PushMessage("Hello with a delay")
			assert.Nil(t, err)
		})

		err := m.ReceiveWithTimeout(20 * time.Millisecond, func(msg interface{}) bool {
			return false
		}, nil)
		assert.NotNil(t, err)
		assert.Equal(t, ErrMailboxReceiveTimeout, err)

		wg.Wait()
	})

	t.Run("timeout reset after reading user's new message", func(t *testing.T) {
		m := NewQueueMailbox(2, 2, 0, DefaultGoSchedulerInterval)

		wg := sync.WaitGroup{}
		wg.Add(2)

		timeout := 100 * time.Millisecond

		// this message is supposed to never get delivered because the delay is greater
		// than the receive timeout
		time.AfterFunc(110 * time.Millisecond, func() {
			defer wg.Done()
			err := m.PushMessage("Hello agian")
			assert.Nil(t, err)
		})

		// but the mailbox is obviously going to receive this message so its timeout
		// checking start-time for will be reset; so both messages should get delivered
		time.AfterFunc(80 * time.Millisecond, func() {
			defer wg.Done()
			err := m.PushMessage("Hello with a delay")
			assert.Nil(t, err)
		})

		i := 0
		err := m.ReceiveWithTimeout(timeout, func(msg interface{}) bool {
			i++
			if i == 2 {return false}
			return true
		}, nil)
		// both messages should get delivered so our message handler is going to return
		// true and exit before the timeout gets triggered.
		assert.Nil(t, err)

		wg.Wait()
	})

	t.Run("timeout reset after reading system's new message", func(t *testing.T) {
		m := NewQueueMailbox(2, 2, 0, DefaultGoSchedulerInterval)

		wg := sync.WaitGroup{}
		wg.Add(2)

		timeout := 100 * time.Millisecond

		// this message is supposed to never get delivered because the delay is greater
		// than the receive timeout
		time.AfterFunc(110 * time.Millisecond, func() {
			defer wg.Done()
			err := m.PushSystemMessage("Hello agian")
			assert.Nil(t, err)
		})

		// but the mailbox is obviously going to receive this message so its timeout
		// checking start-time for will be reset; so both messages should get delivered
		time.AfterFunc(80 * time.Millisecond, func() {
			defer wg.Done()
			err := m.PushSystemMessage("Hello with a delay")
			assert.Nil(t, err)
		})

		i := 0
		err := m.ReceiveWithTimeout(timeout, nil, func(msg interface{}) bool {
			i++
			if i == 2 {return false}
			return true
		})
		// both messages should get delivered so our message handler is going to return
		// true and exit before the timeout gets triggered.
		assert.Nil(t, err)

		wg.Wait()
	})
}

func TestQueueMailboxPush(t *testing.T) {
	m := NewQueueMailbox(2, 2, 10 * time.Millisecond, DefaultGoSchedulerInterval)

	var err error

	for i := 0; i < 2; i++ {
		err = m.PushMessage(fmt.Sprintf("user msg #%d", i))
		assert.Nil(t, err)
		err = m.PushSystemMessage(fmt.Sprintf("sys msg #%d", i))
		assert.Nil(t, err)
	}

	var expectedError = ErrMailboxEnqueueTimeout
	var disposed bool

ErrPoint:
	err = m.PushMessage(fmt.Sprint("user msg #2"))
	if !assert.NotNil(t, err) {return}
	assert.Equal(t, expectedError, err)

	err = m.PushSystemMessage(fmt.Sprint("sys msg #2"))
	if !assert.NotNil(t, err) {return}
	assert.Equal(t, expectedError, err)

	if !disposed {
		m.Dispose()
		disposed = true
		expectedError = ErrMailboxClosed
		goto ErrPoint
	}

	return
}

func TestQueueMailbox_Dispose(t *testing.T) {
	msgHandler := func(msg interface{}) bool {
		return false
	}

	t.Run("disposed", func(t *testing.T) {
		m := NewQueueMailbox(2, 2, 0, DefaultGoSchedulerInterval)

		assert.Equal(t, uint32(0), m.disposed)
		m.Dispose()
		assert.Equal(t, uint32(1), m.disposed)

		err := m.Receive(msgHandler, msgHandler)
		assert.NotNil(t, err)
		assert.Equal(t, ErrMailboxClosed, err)
	})

	t.Run("receive user msg on disposed mailbox", func(t *testing.T) {
		m := NewQueueMailbox(2, 2, 10 * time.Millisecond, DefaultGoSchedulerInterval)

		err := m.PushMessage("Hello dear user")
		assert.Nil(t, err)

		m.userMsgQueue.Dispose()

		err = m.Receive(msgHandler, msgHandler)
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMailboxClosed, err)
	})

	t.Run("receive sys msg on disposed mailbox", func(t *testing.T) {
		m := NewQueueMailbox(2, 2, 10 * time.Millisecond, DefaultGoSchedulerInterval)

		err := m.PushSystemMessage("Hello dear admin")
		assert.Nil(t, err)

		m.sysMsgQueue.Dispose()

		err = m.Receive(msgHandler, msgHandler)
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMailboxClosed, err)
	})
}

func TestNewQueueMailbox(t *testing.T) {
	m := NewQueueMailbox(1, 1, 0, DefaultGoSchedulerInterval)

	assert.Equal(t, uint64(2), m.userMsgQueue.Cap())
	assert.Equal(t, uint64(2), m.sysMsgQueue.Cap())
}