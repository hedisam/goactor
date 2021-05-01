package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/process"
	"github.com/hedisam/goactor/sysmsg"
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
	t.Run("calling dispose function", func(t *testing.T) {
		actor, dispose := NewParentActor(nil)
		assert.NotNil(t, actor)
		assert.NotNil(t, dispose)
		dispose()

		err := Send(actor.Self(), nil)
		assert.NotNil(t, err)

		err = actor.ReceiveWithTimeout(time.Nanosecond * 100, func(message interface{}) (loop bool) {
			return false
		})
		assert.NotNil(t, err)
	})

	t.Run("sending & receiving msg", func(t *testing.T) {
		actor, dispose := NewParentActor(nil)
		assert.NotNil(t, actor)
		assert.NotNil(t, dispose)
		defer dispose()

		msg := "Hi"
		var received interface{}

		err := Send(actor.Self(), msg)
		if !assert.Nil(t, err) {return}

		err = actor.ReceiveWithTimeout(time.Nanosecond * 100, func(message interface{}) (loop bool) {
			received = message
			return false
		})
		assert.Nil(t, err)
		assert.Equal(t, msg, received)
	})

	t.Run("`defer dispose()` should catch panics", func(t *testing.T) {
		actor, dispose := NewParentActor(nil)
		assert.NotNil(t, actor)
		assert.NotNil(t, dispose)

		monitor, _ := setupActor(nil)
		assert.NotNil(t, monitor)

		err := monitor.Monitor(actor.Self())
		assert.Nil(t, err)

		defer func() {
			// the panic should already have been caught by the actor's dispose method
			r := recover()
			assert.Nil(t, r)

			err = monitor.ReceiveWithTimeout(time.Millisecond, func(message interface{}) (loop bool) {
				// the monitor should get notified about the actor's panic
				assert.NotNil(t, message)
				assert.IsType(t, sysmsg.AbnormalExit{}, message)
				return false
			})
			assert.Nil(t, err)
		}()

		defer dispose()

		// parent actors do not spawn an exclusive goroutine, therefore this current routine is
		// our parent's routine; so let's panic here
		panic("This is a parent actor wishing to panic!")
	})
}

func TestSpawn(t *testing.T) {

	t.Run("actor's fn invoked", func(t *testing.T) {
		fnChan := make(chan struct{})
		fn := func(a *Actor) {
			fnChan <- struct{}{}
		}

		pid := Spawn(fn, DefaultChanMailbox)
		assert.NotNil(t, pid)

		invoked := false
		select {
		case <-fnChan:
			invoked = true
		case <-time.After(10 * time.Nanosecond):
		}
		assert.True(t, invoked)
	})

	t.Run("sending & receiving msg on our spawned actor", func(t *testing.T) {
		var received []interface{}
		fnChan := make(chan struct{})
		fn := func(a *Actor) {
			_ = a.ReceiveWithTimeout(time.Millisecond*10, func(message interface{}) (loop bool) {
				received = append(received, message)
				return true
			})
			close(fnChan)
		}

		pid := Spawn(fn, nil)

		n := 100
		numRoutines := 100
		for i := 0; i < numRoutines; i++ {
			go func() {
				for k := 0; k < n; k++ {
					err := Send(pid, k)
					assert.Nil(t, err)
				}
			}()
		}

		<-fnChan

		assert.Equal(t, n * numRoutines, len(received))
	})

	t.Run("spawned actor should survive panics", func(t *testing.T) {
		fn := func(a *Actor) {
			_ = a.Receive(func(message interface{}) (loop bool) {
				panic(message)
			})
		}

		pid := Spawn(fn, nil)
		assert.NotNil(t, pid)

		// monitoring the actor to make sure we receive the exit message caused by panic
		parent, dispose := NewParentActor(nil)
		assert.NotNil(t, parent)
		assert.NotNil(t, dispose)
		defer dispose()

		err := parent.Monitor(pid)
		if !assert.Nil(t, err) {return}

		err = Send(pid, "asking the actor to panic")
		if !assert.Nil(t, err) {return}

		err = parent.ReceiveWithTimeout(time.Millisecond*10, func(message interface{}) (loop bool) {
			assert.NotNil(t, message)
			assert.IsType(t, sysmsg.AbnormalExit{}, message)
			return false
		})
		// timeout should not get triggered because we're supposed to receive the exit message
		// emitted by the actor that had a panic
		assert.Nil(t, err)
	})
}


















