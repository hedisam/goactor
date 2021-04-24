package mailbox

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestChanMailbox_Receive(t *testing.T) {
	var length = 5
	m := NewChanMailbox(length, length, 10 * time.Millisecond)

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
		m := NewChanMailbox(2, 2, 10 * time.Millisecond)

		err := m.ReceiveWithTimeout(20 * time.Millisecond, func(msg interface{}) bool {
			return false
		}, nil)
		assert.NotNil(t, err)
		assert.Equal(t, ErrMailboxReceiveTimeout, err)
	})

	t.Run("timeout reset after reading user's new message", func(t *testing.T) {
		m := NewChanMailbox(2, 2, 10 * time.Millisecond)
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
		// checking start-time will be reset; so both messages should get delivered
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
		m := NewChanMailbox(2, 2, 10 * time.Millisecond)
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
		// checking start-time will be reset; so both messages should get delivered
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

func TestChanMailbox_Push(t *testing.T) {
	testFunc := func(t *testing.T, n int, m *chanMailbox) {
		var err error
		for i := 0; i < n; i++ {
			err = m.PushMessage(i)
			assert.Nil(t, err)
			err = m.PushSystemMessage(i)
			assert.Nil(t, err)
		}

		err = m.PushMessage(n)
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMailboxEnqueueTimeout, err)

		err = m.PushSystemMessage(n)
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMailboxEnqueueTimeout, err)
	}

	for i := 0; i < 20; i++ {
		m := NewChanMailbox(i, i, 3 * time.Millisecond)
		testFunc(t, i, m)
	}
}

func TestChanMailbox_Dispose(t *testing.T) {
	msgHandler := func(msg interface{}) bool {
		return false
	}

	m := NewChanMailbox(1, 1, 0)
	m.Dispose()

	t.Run("calling dispose multiple times", func(t *testing.T) {
		m.Dispose()
		m.Dispose()

	})

	t.Run("push msg on disposed", func(t *testing.T) {
		err := m.PushMessage("Hello")
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMailboxClosed, err)

		err = m.PushSystemMessage("Hi")
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMailboxClosed, err)
	})

	t.Run("receive on disposed mailbox", func(t *testing.T) {
		err := m.Receive(msgHandler, msgHandler)
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMailboxClosed, err)

		err = m.ReceiveWithTimeout(1, msgHandler, msgHandler)
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMailboxClosed, err)
	})

	t.Run("push and receive on disposed mailbox", func(t *testing.T) {
		m := NewChanMailbox(2, 2, 0)
		err := m.PushMessage("Hi")
		if !assert.Nil(t, err) {return}

		m.Dispose()

		err = m.Receive(func(msg interface{}) bool {
			return false
		}, nil)
		if !assert.NotNil(t, err) {return }
		assert.Equal(t, ErrMailboxClosed, err)
	})
}