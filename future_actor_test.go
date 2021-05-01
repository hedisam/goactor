package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestFutureActor_Receive(t *testing.T) {
	future := NewFutureActor()
	if !assert.NotNil(t, future) {return}

	var received interface{}

	msg := "Hello, now you should dispose"
	err := Send(future.Self(), msg)
	if !assert.Nil(t, err) {return}

	err = future.Receive(func(message interface{}) (loop bool) {
		received = message
		return false
	})
	if !assert.Nil(t, err) {return}
	assert.Equal(t, msg, received)

	// the receive method has returned so we expect future's mailbox to be disposed

	msg = "this msg should fail"
	err = Send(future.Self(), msg)
	assert.NotNil(t, err)
}

func TestFutureActor_ReceiveWithTimeout(t *testing.T) {
	future := NewFutureActor()
	if !assert.NotNil(t, future) {return}

	var received []interface{}
	timeout := 10 * time.Millisecond

	msg := "this message should get delivered"
	err := Send(future.Self(), msg)
	assert.Nil(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	time.AfterFunc(timeout * 2, func() {
		defer wg.Done()
		msg := "this message will not get delivered"
		err = Send(future.Self(), msg)
		assert.NotNil(t, err)
	})

	err = future.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
		received = append(received, message)
		return true
	})
	assert.NotNil(t, err)
	assert.Equal(t, 1, len(received))

	wg.Wait()
}

func TestFutureActor_systemMessageHandler(t *testing.T) {
	// all system messages sent a future actor must be delivered to the user's mailbox
	future := NewFutureActor()
	if !assert.NotNil(t, future) {return}

	var received interface{}
	msg := "This is a system message"
	err := intlpid.SendSystemMessage(future.Self().InternalPID(), msg)
	if !assert.Nil(t, err) {return}

	err = future.ReceiveWithTimeout(10 * time.Millisecond, func(message interface{}) (loop bool) {
		received = message
		return false
	})
	assert.Nil(t, err)

	if !assert.NotNil(t, received) {return}
	assert.Equal(t, msg, received)
}

func TestFutureActor_Self(t *testing.T) {
	future := NewFutureActor()
	if !assert.NotNil(t, future) {return}

	self := future.Self()
	if !assert.NotNil(t, self) {return}
	assert.NotNil(t, self.InternalPID())
}
