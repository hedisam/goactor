package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/process"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	t.Run("sending to a nil pid", func(t *testing.T) {
		err := Send(nil, "this msg will fail")
		assert.NotNil(t, err)
		assert.Equal(t, ErrSendNilPID, err)
	})

	t.Run("sending msg to a supervisor", func(t *testing.T) {
		iPID := intlpid.NewLocalPID(nil, nil, true, func(){})
		pid := p.ToPID(iPID)
		if !assert.NotNil(t, pid) {return}

		err := Send(pid, "we can't directly send msg to a supervisor")
		assert.NotNil(t, err)
		assert.Equal(t, ErrSendToSupervisor, err)
	})

	t.Run("send message to a healthy actor", func(t *testing.T) {
		actor, pid := setupActor(nil)
		if !assert.NotNil(t, pid) {return}

		msg := "testing Send"
		err := Send(pid, msg)
		assert.Nil(t, err)

		var received interface{}
		err = actor.ReceiveWithTimeout(time.Nanosecond * 10, func(message interface{}) (loop bool) {
			received = message
			return false
		})
		assert.Nil(t, err)
		assert.Equal(t, msg, received)
	})
}

func TestSendNamed(t *testing.T) {
	actor, pid := setupActor(DefaultChanMailbox)
	if !assert.NotNil(t, actor) {return}
	if !assert.NotNil(t, pid) {return}

	var name = "my_actor"

	t.Run("send to a registered actor", func(t *testing.T) {
		process.Register(name, pid)
		msg := "Hi you named actor"
		var received interface{}

		err := SendNamed(name, msg)
		if !assert.Nil(t, err) {return}


		err = actor.ReceiveWithTimeout(10 * time.Nanosecond, func(message interface{}) (loop bool) {
			received = message
			return false
		})
		if !assert.Nil(t, err) {return}
		assert.NotNil(t, received)
		assert.Equal(t, msg, received)
	})

	t.Run("send to a non registered actor(name)", func(t *testing.T) {
		process.Unregister(name)

		msg := "This will fail"

		err := SendNamed(name, msg)
		assert.NotNil(t, err)
		assert.Equal(t, ErrSendNameNotFound, err)
	})
}

func TestNewParentActor(t *testing.T) {

}




















