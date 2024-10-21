package intprocess

import (
	"slices"
	"sync"
)

// relationType represents the relationType type.
type relationType int

const (
	// relationLinked represents an actor with a bidirectional relationship.
	relationLinked relationType = iota
	// relationMonitored represents an actor that is being monitored.
	relationMonitored
	// relationMonitor represents an actor that is monitoring the actor.
	relationMonitor
)

// relations manages relations between actors.
type relations struct {
	mu                sync.RWMutex
	idToRelationTypes map[string][]relationType
	idToPID           map[string]PID
}

func newRelations() *relations {
	return &relations{
		idToRelationTypes: make(map[string][]relationType),
		idToPID:           make(map[string]PID),
	}
}

// Add adds a new relationType for the provided PID.
func (m *relations) Add(pid PID, rel relationType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rels := m.idToRelationTypes[pid.Ref()]
	if idx := slices.Index(rels, rel); idx != -1 {
		// relation already exists
		return
	}
	m.idToRelationTypes[pid.Ref()] = append(rels, rel)
	m.idToPID[pid.Ref()] = pid
}

// Remove removes a specific type of relationType for the provided process ID.
func (m *relations) Remove(id string, rel relationType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rels := m.idToRelationTypes[id]
	idx := slices.Index(rels, rel)
	switch {
	case idx == -1:
		return
	case idx == len(rels)-1:
		rels = rels[:idx]
	default:
		rels = append(rels[:idx], rels[idx+1:]...)
	}

	if len(rels) == 0 {
		m.purge(id)
		return
	}

	m.idToRelationTypes[id] = rels
}

// Purge purges all the relations for the provided process ID.
func (m *relations) Purge(id string) {
	m.mu.Lock()
	m.purge(id)
	m.mu.Unlock()
}

func (m *relations) purge(id string) {
	delete(m.idToRelationTypes, id)
	delete(m.idToPID, id)
}

// TypeToRelatedPIDs returns a map of relation type to a slice PID.
func (m *relations) TypeToRelatedPIDs() map[relationType][]PID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	typeToPIDs := make(map[relationType][]PID)
	for id, types := range m.idToRelationTypes {
		pid := m.idToPID[id]
		for typ := range slices.Values(types) {
			typeToPIDs[typ] = append(typeToPIDs[typ], pid)
		}
	}
	return typeToPIDs
}