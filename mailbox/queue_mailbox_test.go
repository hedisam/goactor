package mailbox

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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

	var expectedError = ErrMailboxTimeout
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
