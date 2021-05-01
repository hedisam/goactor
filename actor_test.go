package goactor

import (
	"github.com/hedisam/goactor/mailbox"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/sysmsg"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestActor_TrapExit(t *testing.T) {
	actor, _ := getActorForTest(t)

	assert.False(t, actor.TrapExit())

	actor.SetTrapExit(true)
	assert.True(t, actor.TrapExit())

	actor.SetTrapExit(false)
	assert.False(t, actor.TrapExit())
}

func TestActor_Self(t *testing.T) {
	actor, pid := getActorForTest(t)

	self := actor.Self()
	assert.NotNil(t, self)
	assert.Equal(t, pid, self)
}

func TestActor_Receive(t *testing.T) {
	actor, pid := setupActor(DefaultChanMailbox)
	if !assert.NotNil(t, pid) || !assert.NotNil(t, actor) {return}

	err := Send(pid, "Hello")
	assert.Nil(t, err)

	err = actor.Receive(func(message interface{}) (loop bool) {
		if !assert.NotNil(t, message) {return false}
		assert.Equal(t, "Hello", message)
		return false
	})
	assert.Nil(t, err)

	messages := []interface{}{"Hi", 2, "Hidayat"}
	for _, msg := range messages {
		err = Send(pid, msg)
		if !assert.Nil(t, err) {return}
	}

	i := 0
	err = actor.Receive(func(message interface{}) (loop bool) {
		if !assert.NotNil(t, messages) {return false}
		if !assert.Equal(t, messages[i], message) {return false}
		i++
		if i >= len(messages) {
			return false
		}
		return true
	})
	assert.Nil(t, err)
	assert.Equal(t, len(messages), i)

	actor.mailbox.Dispose()

	err = actor.Receive(func(message interface{}) (loop bool) {
		t.Errorf("expected to not receive any messages, but got: %v", message)
		return false
	})
	if !assert.NotNil(t, err) {return}
	// todo: receive should return its own err
	assert.Equal(t, mailbox.ErrMailboxClosed, err)
}

func TestActor_ReceiveWithTimeout(t *testing.T) {
	actor, pid := setupActor(DefaultChanMailbox)
	if !assert.NotNil(t, pid) || !assert.NotNil(t, actor) {return}

	err := Send(pid, "Hello")
	if !assert.Nil(t, err) {return}

	err = actor.ReceiveWithTimeout(time.Millisecond * 10, func(message interface{}) (loop bool) {
		if !assert.NotNil(t, message) {return}
		assert.Equal(t, "Hello", message)
		return false
	})
	if !assert.Nil(t, err) {return}

	time.AfterFunc(time.Millisecond * 100, func() {
		err = Send(pid, "Hi with delay")
		assert.Nil(t, err)
	})

	err = actor.ReceiveWithTimeout(time.Millisecond * 10, func(message interface{}) (loop bool) {
		t.Errorf("timeout expected to get triggered before receiving this message: %v", message)
		return false
	})
	if !assert.NotNil(t, err) {return}
	assert.Equal(t, mailbox.ErrMailboxReceiveTimeout, err)

	actor.mailbox.Dispose()

	err = actor.ReceiveWithTimeout(time.Millisecond * 10, func(message interface{}) (loop bool) {
		t.Errorf("expected to receive no messages due to having a closed mailbox, but received: %v", message)
		return false
	})
	if !assert.NotNil(t, err) {return}
	assert.Equal(t, mailbox.ErrMailboxClosed, err)
}

func TestActor_LinkUnlink(t *testing.T) {
	actor1, pid1 := setupActor(DefaultChanMailbox)
	actor2, pid2 := setupActor(DefaultChanMailbox)
	if !assert.NotNil(t, actor1) || !assert.NotNil(t, actor2) ||
		!assert.NotNil(t, pid1) || !assert.NotNil(t, pid2) {
		return
	}

	it := actor1.relationManager.LinkedActors()
	assert.False(t, it.HasNext())

	err := actor1.Link(pid2)
	assert.Nil(t, err)

	it = actor1.relationManager.LinkedActors()
	if !assert.True(t, it.HasNext()) {return}
	linkedToActor1 := it.Value()
	assert.NotNil(t, linkedToActor1)
	assert.Equal(t, pid2.InternalPID(), linkedToActor1)

	it = actor2.relationManager.LinkedActors()
	if !assert.True(t, it.HasNext()) {return}
	linkedToActor2 := it.Value()
	assert.NotNil(t, linkedToActor2)
	assert.Equal(t, pid1.InternalPID(), linkedToActor2)

	// we can not link the same actor twice
	err = actor1.Link(pid2)
	assert.Nil(t, err)

	it = actor1.relationManager.LinkedActors()
	if !assert.True(t, it.HasNext()) {return}
	assert.Equal(t, pid2.InternalPID(), it.Value())
	if !assert.False(t, it.HasNext()) {return}

	it = actor2.relationManager.LinkedActors()
	if !assert.True(t, it.HasNext()) {return}
	assert.Equal(t, pid1.InternalPID(), it.Value())
	assert.False(t, it.HasNext())

	actor3, pid3 := setupActor(DefaultChanMailbox)
	if !assert.NotNil(t, actor3) || !assert.NotNil(t, pid3) {return}

	err = actor1.Link(pid3)
	if !assert.Nil(t, err) {return}

	it = actor1.relationManager.LinkedActors()
	if !assert.True(t, it.HasNext()) {return}
	secondLinkedActor := it.Value()
	if reflect.DeepEqual(secondLinkedActor, pid2.InternalPID()) {
		if !assert.True(t, it.HasNext()) {return}
		secondLinkedActor = it.Value()
		if !assert.Equal(t, pid3.InternalPID(), secondLinkedActor) {return}
	} else if reflect.DeepEqual(secondLinkedActor, pid3.InternalPID()) {
		if !assert.True(t, it.HasNext()) {return}
		secondLinkedActor = it.Value()
		if !assert.Equal(t, pid2.InternalPID(), secondLinkedActor) {return}
	} else {
		t.Errorf("unknown actor linked to our actor: expected(%v or %v), got: %v", pid2.InternalPID(), pid3.InternalPID(), secondLinkedActor)
		return
	}
	if !assert.False(t, it.HasNext()) {return }

	err = actor1.Link(nil)
	if !assert.NotNil(t, err) {return}
	assert.Equal(t, ErrLinkNilTargetPID, err)

	err = actor1.Unlink(nil)
	if !assert.NotNil(t, err) {return}
	assert.Equal(t, ErrUnlinkNilTargetPID, err)

	err = actor1.Unlink(pid2)
	if !assert.Nil(t, err) {return}

	it = actor1.relationManager.LinkedActors()
	if !assert.True(t, it.HasNext()) {return }
	assert.Equal(t, pid3.InternalPID(), it.Value())
	assert.False(t, it.HasNext())

	it = actor2.relationManager.LinkedActors()
	assert.False(t, it.HasNext())

	err = actor1.Unlink(pid3)
	assert.Nil(t, err)

	it = actor1.relationManager.LinkedActors()
	assert.False(t, it.HasNext())



	actor2.relationManager.Dispose()

	err = actor1.Link(pid2)
	if !assert.NotNil(t, err) {return}

	actor1.relationManager.Dispose()

	err = actor1.Link(pid3)
	if !assert.NotNil(t, err) {return}

	actor4, pid4 := setupActor(DefaultChanMailbox)
	if !assert.NotNil(t, actor4) || !assert.NotNil(t, pid4) {return}

	actor3.relationManager.Dispose()

	err = actor4.Unlink(pid3)
	if !assert.NotNil(t, err) {return}

	actor4.relationManager.Dispose()

	actor5, pid5 := setupActor(DefaultChanMailbox)
	if !assert.NotNil(t, actor5) || !assert.NotNil(t, pid5) {return}

	err = actor4.Unlink(pid5)
	assert.NotNil(t, err)
}

func TestActor_MonitorDemonitor(t *testing.T) {
	t.Run("(de)monitor another actor", func(t *testing.T) {
		actor1, pid1 := getActorForTest(t)
		actor2, pid2 := getActorForTest(t)

		err := actor1.Monitor(pid2)
		if !assert.Nil(t, err) {return}

		it := actor2.relationManager.MonitorActors()
		assert.True(t, it.HasNext())
		assert.Equal(t, pid1.InternalPID(), it.Value())
		assert.False(t, it.HasNext())

		err = actor1.Demonitor(pid2)
		if !assert.Nil(t, err) {return}

		it = actor2.relationManager.MonitorActors()
		assert.False(t, it.HasNext())
	})

	t.Run("monitor an already monitored actor", func(t *testing.T) {
		actor1, pid1 := getActorForTest(t)
		actor2, pid2 := getActorForTest(t)

		err := actor1.Monitor(pid2)
		if !assert.Nil(t, err) {return}

		// this should not add anything to the list of monitored actors
		err = actor1.Monitor(pid2)
		if !assert.Nil(t, err) {return}

		it := actor2.relationManager.MonitorActors()
		assert.True(t, it.HasNext())
		assert.Equal(t, pid1.InternalPID(), it.Value())
		assert.False(t, it.HasNext())
	})

	t.Run("demonitor a non-monitored actor", func(t *testing.T) {
		actor1, _ := getActorForTest(t)
		_, pid2 := getActorForTest(t)

		err := actor1.Monitor(pid2)
		if !assert.Nil(t, err) {return}
	})

	t.Run("(de)monitor nil pid", func(t *testing.T) {
		actor1, _ := getActorForTest(t)

		err := actor1.Monitor(nil)
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrMonitorNilTargetPID, err)

		err = actor1.Demonitor(nil)
		if !assert.NotNil(t, err) {return}
		assert.Equal(t, ErrDemonitorNilTargetPID, err)
	})

	t.Run("(de)monitor disposed actor", func(t *testing.T) {
		actor1, _ := getActorForTest(t)
		actor2, pid2 := getActorForTest(t)

		actor2.relationManager.Dispose()

		err := actor1.Monitor(pid2)
		assert.NotNil(t, err)

		err = actor1.Demonitor(pid2)
		assert.NotNil(t, err)
	})

	t.Run("(de)monitor while the actor itself is disposed", func(t *testing.T) {
		actor1, _ := getActorForTest(t)
		_, pid2 := getActorForTest(t)

		actor1.relationManager.Dispose()

		err := actor1.Monitor(pid2)
		assert.NotNil(t, err)

		err = actor1.Demonitor(pid2)
		assert.NotNil(t, err)
	})
}

func TestActor_Context(t *testing.T) {
	actor, _ := getActorForTest(t)

	ctx := actor.Context()
	assert.NotNil(t, ctx)

	actor.ctxCancel()
	canceled := false
	select {
	case <-ctx.Done():
		canceled = true
	case <-time.After(10 * time.Nanosecond):
	}

	assert.True(t, canceled)
}

func TestActor_systemMessageHandlerNormalExit(t *testing.T) {
	actor, _ := getActorForTest(t)
	var msgReceived bool

	actor.msgHandler = func(message interface{}) (loop bool) {
		msgReceived = true
		return false
	}

	t.Run("without any relation", func(t *testing.T) {
		msg := sysmsg.NewNormalExitMsg(nil, nil)

		loop := actor.systemMessageHandler(msg)

		// msg's sender is not linked or monitored by this actor, so the msg should not
		// get forwarded to the user's msgHandler
		assert.False(t, msgReceived)
		// systemMessageHandler return true to signal the mailbox to continue listening
		assert.True(t, loop)
	})

	t.Run("msg from linked actor", func(t *testing.T) {
		_, pid2 := getActorForTest(t)
		msg := sysmsg.NewNormalExitMsg(pid2.InternalPID(), nil)
		msgReceived = false

		err := actor.Link(pid2)
		if !assert.Nil(t, err) {return}

		loop := actor.systemMessageHandler(msg)

		// the msg should get forwarded to the user's msgHandler
		assert.True(t, msgReceived)
		// since our msgHandler sets msgReceived to true and then returns false, our
		// systemMessageHandler should return false, too.
		assert.False(t, loop)
	})

	t.Run("msg from a monitored actor", func(t *testing.T) {
		_, pid2 := getActorForTest(t)
		msg := sysmsg.NewNormalExitMsg(pid2.InternalPID(), nil)
		msgReceived = false

		err := actor.Monitor(pid2)
		if !assert.Nil(t, err) {return}

		loop := actor.systemMessageHandler(msg)

		assert.True(t, msgReceived)
		assert.False(t, loop)
	})
}

func TestActor_systemMessageHandlerAbnormalMessage(t *testing.T) {
	actor, _ := getActorForTest(t)
	var msgReceived bool

	actor.msgHandler = func(message interface{}) (loop bool) {
		msgReceived = true
		return false
	}

	t.Run("without any relation", func(t *testing.T) {
		msg := sysmsg.NewAbnormalExitMsg(nil,nil, nil)

		loop := actor.systemMessageHandler(msg)

		// msg's sender is not linked or monitored by this actor, so the msg should not
		// get forwarded to the user's msgHandler
		assert.False(t, msgReceived)
		// systemMessageHandler return true to signal the mailbox to continue listening
		assert.True(t, loop)
	})

	t.Run("sys msg from monitored actor", func(t *testing.T) {
		_, pid2 := getActorForTest(t)
		msg := sysmsg.NewAbnormalExitMsg(pid2.InternalPID(), nil, nil)
		msgReceived = false

		err := actor.Monitor(pid2)
		if !assert.Nil(t, err) {return}

		loop := actor.systemMessageHandler(msg)

		assert.True(t, msgReceived)
		assert.False(t, loop)
	})

	t.Run("sys msg from linked actor & trapping exit messages", func(t *testing.T) {
		_, pid2 := getActorForTest(t)
		msg := sysmsg.NewAbnormalExitMsg(pid2.InternalPID(), nil, nil)
		msgReceived = false

		actor.SetTrapExit(true)

		err := actor.Link(pid2)
		if !assert.Nil(t, err) {return}

		loop := actor.systemMessageHandler(msg)

		// the message should get forwarded to the user's msgHandler
		assert.True(t, msgReceived)
		assert.False(t, loop)
	})

	t.Run("sys msg from linked actor without trapping exit messages", func(t *testing.T) {
		_, pid2 := getActorForTest(t)
		msg := sysmsg.NewAbnormalExitMsg(pid2.InternalPID(), "just testing", nil)
		msgReceived = false

		actor.SetTrapExit(false)

		err := actor.Link(pid2)
		if !assert.Nil(t, err) {return}

		// as a reason/message for the panic
		// trapping exit messages, it should simply just panic with the received message
		// since our actor is linked to the one sending the abnormal message and also not
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Equal(t, msg, r)
		}()

		_ = actor.systemMessageHandler(msg)
	})
}

func TestActor_dispose(t *testing.T) {
	actor, pid := getActorForTest(t)
	monitorActor, _ := getActorForTest(t)
	linkedActor, lPID := getActorForTest(t)

	var timeout = 10 * time.Millisecond

	err := monitorActor.Monitor(pid)
	if !assert.Nil(t, err) {return}
	err = actor.Link(lPID)
	if !assert.Nil(t, err) {return}

	// trapping exit messages so we can check if they have received the corresponding system
	// message from our actor
	monitorActor.SetTrapExit(true)
	linkedActor.SetTrapExit(true)

	t.Run("testing actor's shutdown", func(t *testing.T) {
		actor, pid := getActorForTest(t)

		actor.dispose()

		// the mailbox should be closed
		err := Send(pid, "this msg should fail")
		assert.NotNil(t, err)

		// we should not be able to link nor monitor any other actors
		_, pid2 := getActorForTest(t)
		err = actor.Link(pid2)
		assert.NotNil(t, err)

		// the actor's context should get canceled
		canceled := false
		select {
		case <-actor.Context().Done():
			canceled = true
		case <-time.After(10 * time.Millisecond):
		}
		assert.True(t, canceled)
	})

	t.Run("when the actor finishes its work normally", func(t *testing.T) {
		actor.dispose()

		err = monitorActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
			assert.NotNil(t, message)
			assert.IsType(t, sysmsg.NormalExit{}, message)
			return false
		})
		assert.Nil(t, err) // no timeout error, so the monitor actor has received the exit message

		err = linkedActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
			assert.NotNil(t, message)
			assert.IsType(t, sysmsg.NormalExit{}, message)
			return false
		})
		assert.Nil(t, err) // no timeout error, so the linked actor has received the exit msg
	})

	t.Run("exited due to an abnormal exit message", func(t *testing.T) {
		defer func() {
			err = monitorActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
				assert.NotNil(t, message)
				assert.IsType(t, sysmsg.AbnormalExit{}, message)
				return false
			})
			assert.Nil(t, err)

			err = linkedActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
				assert.NotNil(t, message)
				assert.IsType(t, sysmsg.AbnormalExit{}, message)
				return false
			})
			assert.Nil(t, err)
		}()

		defer actor.dispose()
		// panic-ing with an AbnormalExit to simulate the situation
		_, exitMsgSenderPID := getActorForTest(t)
		panic(sysmsg.NewAbnormalExitMsg(exitMsgSenderPID.InternalPID(), "just testing", nil))
	})

	t.Run("exited because of receiving a NormalExit msg", func(t *testing.T) {
		defer func() {
			err = monitorActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
				assert.NotNil(t, message)
				assert.IsType(t, sysmsg.NormalExit{}, message)
				return false
			})
			assert.Nil(t, err)

			err = linkedActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
				assert.NotNil(t, message)
				assert.IsType(t, sysmsg.NormalExit{}, message)
				return false
			})
			assert.Nil(t, err)
		}()

		defer actor.dispose()
		_, exitMsgSenderPID := getActorForTest(t)
		panic(sysmsg.NewNormalExitMsg(exitMsgSenderPID.InternalPID(), nil))
	})

	t.Run("exited because of panic-ing due to an unknown err", func(t *testing.T) {
		defer func() {
			err = monitorActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
				assert.NotNil(t, message)
				return false
			})
			assert.Nil(t, err)

			err = linkedActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
				assert.NotNil(t, message)
				return false
			})
			assert.Nil(t, err)
		}()

		defer actor.dispose()
		panic("unknown situation")
	})

	t.Run("when exitmsg sender is in the list of notifiable actors", func(t *testing.T) {
		defer func() {
			err = monitorActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
				assert.NotNil(t, message)
				assert.IsType(t, sysmsg.AbnormalExit{}, message)
				return false
			})
			assert.Nil(t, err)

			err = linkedActor.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
				return false
			})
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "timeout")
		}()

		defer actor.dispose()
		panic(sysmsg.NewAbnormalExitMsg(lPID.InternalPID(), "testing", nil))
	})
}

func getActorForTest(t *testing.T) (*Actor, *p.PID) {
	actor, pid := setupActor(DefaultChanMailbox)

	assert.NotNil(t, actor)
	assert.NotNil(t, pid)

	return actor, pid
}
























