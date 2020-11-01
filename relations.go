package goactor

import (
	"sync"
)

type RelationType int32
type RelationMap map[PID]struct{}

const (
	LinkedRelation RelationType = iota
	MonitoredRelation
	NoRelation
)

type Relations struct {
	linkedActors  RelationMap
	monitorActors RelationMap
	sync.RWMutex
}

func newRelations() *Relations {
	return &Relations{
		linkedActors:  make(RelationMap),
		monitorActors: make(RelationMap),
	}
}

func (r *Relations) RelationType(pid PID) RelationType {
	r.RLock()
	defer r.RUnlock()
	if _, ok := r.linkedActors[pid]; ok {
		return LinkedRelation
	} else if _, ok := r.monitorActors[pid]; ok {
		return MonitoredRelation
	}
	return NoRelation
}

func (r *Relations) Link(to PID) {
	r.Lock()
	defer r.Unlock()
	r.linkedActors[to] = struct{}{}
}

func (r *Relations) Unlink(from PID) {
	r.Lock()
	defer r.Unlock()
	delete(r.linkedActors, from)
}

func (r *Relations) BeMonitored(by PID) {
	r.Lock()
	defer r.Unlock()
	r.monitorActors[by] = struct{}{}
}

func (r *Relations) BeDemonitored(by PID) {
	r.Lock()
	defer r.Unlock()
	delete(r.monitorActors, by)
}

func (r *Relations) LinkedActors() map[PID]struct{} {
	r.RLock()
	defer r.RUnlock()

	linkedActors := make(RelationMap)
	for k, _ := range r.linkedActors {
		linkedActors[k] = struct{}{}
	}
	return linkedActors
}

func (r *Relations) MonitorActors() map[PID]struct{} {
	r.RLock()
	defer r.RUnlock()

	monitorActors := make(RelationMap)
	for k, _ := range r.monitorActors {
		monitorActors[k] = struct{}{}
	}
	return monitorActors
}
