package goactor

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

// relationsManager manages relations between actors.
type relationsManager struct {
	mu            sync.RWMutex
	idToRelations map[string][]relationType
	idToPID       map[string]*PID
}

func newRelationsManager() *relationsManager {
	return &relationsManager{
		idToRelations: make(map[string][]relationType),
		idToPID:       make(map[string]*PID),
	}
}

// Add adds a new relationType for the provided PID.
func (m *relationsManager) Add(pid *PID, rel relationType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rels := m.idToRelations[pid.ID()]
	rels = append(rels, rel)
	slices.Sort(rels)
	rels = slices.Compact(rels)
	m.idToRelations[pid.ID()] = rels
	m.idToPID[pid.ID()] = pid
}

// Remove removes a specific type of relationType for the provided process ID.
func (m *relationsManager) Remove(id string, rel relationType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rels := m.idToRelations[id]
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

	m.idToRelations[id] = rels
}

// Purge purges all the relations for the provided process ID.
func (m *relationsManager) Purge(id string) {
	m.mu.Lock()
	m.purge(id)
	m.mu.Unlock()
}

func (m *relationsManager) purge(id string) {
	delete(m.idToRelations, id)
	delete(m.idToPID, id)
}

// Relations returns all the relations for the provided process ID.
func (m *relationsManager) Relations(id string) []relationType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	relations, _ := m.idToRelations[id]
	return relations
}
