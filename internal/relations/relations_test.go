package relations

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getNewMockPID() intlpid.InternalPID {
	return intlpid.NewMockInternalPID()
}

func TestRelations_RelationType(t *testing.T) {
	rm := NewRelation()

	linkedPID := getNewMockPID()
	monitoredPID := getNewMockPID()
	noRelationPID := getNewMockPID()

	err := rm.AddLink(linkedPID)
	assert.Nil(t, err)
	err = rm.AddMonitored(monitoredPID)
	assert.Nil(t, err)
	// no adding for noRelationPID

	relType := rm.RelationType(linkedPID)
	assert.Equal(t, LinkedRelation, relType)

	relType = rm.RelationType(monitoredPID)
	assert.Equal(t, MonitoredRelation, relType)

	relType = rm.RelationType(noRelationPID)
	assert.Equal(t, NoRelation, relType)

	relType = rm.RelationType(nil)
	assert.Equal(t, NoRelation, relType)
}

func TestRelations_AddRemove(t *testing.T) {
	rm := NewRelation()

	t.Run("nil pid", func(t *testing.T) {
		err := rm.AddLink(nil)
		assert.NotNil(t, err)
		err = rm.RemoveLink(nil)
		assert.NotNil(t, err)

		err = rm.AddMonitor(nil)
		assert.NotNil(t, err)
		err = rm.RemoveMonitor(nil)
		assert.NotNil(t, err)

		err = rm.AddMonitored(nil)
		assert.NotNil(t, err)
		err = rm.RemoveMonitored(nil)
		assert.NotNil(t, err)
	})

	t.Run("linked actors", func(t *testing.T) {
		linkedPID := getNewMockPID()
		err := rm.AddLink(linkedPID)
		assert.Nil(t, err)

		pid, ok := rm.linkedActors[linkedPID.ID()]
		if !assert.True(t, ok) {
			return
		}
		assert.Equal(t, linkedPID, pid)

		err = rm.RemoveLink(linkedPID)
		assert.Nil(t, err)

		pid, ok = rm.linkedActors[linkedPID.ID()]
		assert.False(t, ok)
	})

	t.Run("monitored actors", func(t *testing.T) {
		monitoredPID := getNewMockPID()
		err := rm.AddMonitored(monitoredPID)
		assert.Nil(t, err)

		pid, ok := rm.monitoredActors[monitoredPID.ID()]
		if !assert.True(t, ok) {
			return
		}
		assert.Equal(t, monitoredPID, pid)

		err = rm.RemoveMonitored(monitoredPID)
		assert.Nil(t, err)

		pid, ok = rm.monitoredActors[monitoredPID.ID()]
		assert.False(t, ok)
	})

	t.Run("monitor actors", func(t *testing.T) {
		monitorPID := getNewMockPID()
		err := rm.AddMonitor(monitorPID)
		assert.Nil(t, err)

		pid, ok := rm.monitorActors[monitorPID.ID()]
		if !assert.True(t, ok) {return}
		assert.Equal(t, monitorPID, pid)

		err = rm.RemoveMonitor(monitorPID)
		assert.Nil(t, err)

		pid, ok = rm.monitorActors[monitorPID.ID()]
		assert.False(t, ok)
	})

	t.Run("disposed relation manager", func(t *testing.T) {
		rm.Dispose()
		pid := getNewMockPID()

		err := rm.AddLink(pid)
		assert.NotNil(t, err)
		err = rm.RemoveLink(pid)
		assert.NotNil(t, err)

		err = rm.AddMonitor(pid)
		assert.NotNil(t, err)
		err = rm.RemoveMonitor(pid)
		assert.NotNil(t, err)

		err = rm.AddMonitored(pid)
		assert.NotNil(t, err)
		err = rm.RemoveMonitored(pid)
		assert.NotNil(t, err)
	})
}

func TestRelations_LinkedActors(t *testing.T) {
	rm := NewRelation()

	L := 10
	for i := 0; i < L; i++ {
		pid := getNewMockPID()
		err := rm.AddLink(pid)
		assert.Nil(t, err)
	}

	iterator := rm.LinkedActors()
	i := 0
	for iterator.HasNext() {
		i++
		pid := iterator.Value()
		relType := rm.RelationType(pid)
		if !assert.Equal(t, LinkedRelation, relType) {
			return
		}
	}
	assert.Equal(t, L, i)
}

func TestRelations_MonitorActors(t *testing.T) {
	rm := NewRelation()

	L := 10
	for i := 0; i < L; i++ {
		pid := getNewMockPID()
		err := rm.AddMonitor(pid)
		assert.Nil(t, err)
	}

	iterator := rm.MonitorActors()
	i := 0
	for iterator.HasNext() {
		i++
		pid := iterator.Value()
		relType := rm.RelationType(pid)
		if !assert.Equal(t, MonitorRelation, relType) {
			return
		}
	}
	assert.Equal(t, L, i)
}