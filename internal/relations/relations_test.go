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

	rm.AddLink(linkedPID)
	rm.AddMonitored(monitoredPID)
	// no adding for noRelationPID

	relType := rm.RelationType(linkedPID)
	assert.Equal(t, LinkedRelation, relType)

	relType = rm.RelationType(monitoredPID)
	assert.Equal(t, MonitoredRelation, relType)

	relType = rm.RelationType(noRelationPID)
	assert.Equal(t, NoRelation, relType)
}

func TestRelations_AddRemove(t *testing.T) {
	rm := NewRelation()

	t.Run("linked actors", func(t *testing.T) {
		linkedPID := getNewMockPID()
		rm.AddLink(linkedPID)

		pid, ok := rm.linkedActors[linkedPID.ID()]
		if !assert.True(t, ok) {
			return
		}
		assert.Equal(t, linkedPID, pid)

		rm.RemoveLink(linkedPID)

		pid, ok = rm.linkedActors[linkedPID.ID()]
		if !assert.False(t, ok) {
			return
		}
	})

	t.Run("monitored actors", func(t *testing.T) {
		monitoredPID := getNewMockPID()
		rm.AddMonitored(monitoredPID)

		pid, ok := rm.monitoredActors[monitoredPID.ID()]
		if !assert.True(t, ok) {
			return
		}
		assert.Equal(t, monitoredPID, pid)

		rm.RemoveMonitored(monitoredPID)

		pid, ok = rm.monitoredActors[monitoredPID.ID()]
		if !assert.False(t, ok) {
			return
		}
	})

	t.Run("monitor actors", func(t *testing.T) {
		monitorPID := getNewMockPID()
		rm.AddMonitor(monitorPID)

		pid, ok := rm.monitorActors[monitorPID.ID()]
		if !assert.True(t, ok) {return}
		assert.Equal(t, monitorPID, pid)

		rm.RemoveMonitor(monitorPID)

		pid, ok = rm.monitorActors[monitorPID.ID()]
		if !assert.False(t, ok) {return}
	})
}

func TestRelations_LinkedActors(t *testing.T) {
	rm := NewRelation()

	L := 10
	for i := 0; i < L; i++ {
		pid := getNewMockPID()
		rm.AddLink(pid)
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
		rm.AddMonitor(pid)
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