package goactor

import (
	"github.com/hedisam/goactor/mailbox"
	p "github.com/hedisam/goactor/pid"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

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

func getActorForTest(t *testing.T) (*Actor, *p.PID) {
	actor, pid := setupActor(DefaultChanMailbox)

	assert.NotNil(t, actor)
	assert.NotNil(t, pid)

	return actor, pid
}
























