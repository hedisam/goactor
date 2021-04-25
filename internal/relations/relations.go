package relations

import (
	p "github.com/hedisam/goactor/internal/intlpid"
	"sync"
)

type RelationType int32
type RelationMap map[string]p.InternalPID

const (
	LinkedRelation RelationType = iota
	MonitoredRelation
	MonitorRelation
	NoRelation
)

type Relations struct {
	linkedActors    RelationMap
	monitorActors   RelationMap
	monitoredActors RelationMap
	sync.RWMutex
}

func NewRelation() *Relations {
	return &Relations{
		linkedActors:    make(RelationMap),
		monitorActors:   make(RelationMap),
		monitoredActors: make(RelationMap),
	}
}

func (r *Relations) RelationType(pid p.InternalPID) RelationType {
	r.RLock()
	defer r.RUnlock()
	if _, ok := r.linkedActors[pid.ID()]; ok {
		return LinkedRelation
	} else if _, ok = r.monitoredActors[pid.ID()]; ok {
		return MonitoredRelation
	} else if _, ok = r.monitorActors[pid.ID()]; ok {
		return MonitorRelation
	}
	return NoRelation
}

func (r *Relations) AddLink(to p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	r.linkedActors[to.ID()] = to
}

func (r *Relations) RemoveLink(from p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	delete(r.linkedActors, from.ID())
}

func (r *Relations) AddMonitored(who p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	r.monitoredActors[who.ID()] = who
}

func (r *Relations) RemoveMonitored(who p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	delete(r.monitoredActors, who.ID())
}

func (r *Relations) AddMonitor(by p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	r.monitorActors[by.ID()] = by
}

func (r *Relations) RemoveMonitor(by p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	delete(r.monitorActors, by.ID())
}

func (r *Relations) LinkedActors() *RelationIterator {
	r.RLock()
	defer r.RUnlock()

	linkedActors := make([]p.InternalPID, 0, len(r.linkedActors))
	for _, pid := range r.linkedActors {
		linkedActors = append(linkedActors, pid)
	}

	iterator := &RelationIterator{
		data:   linkedActors,
		pos:    0,
		length: len(linkedActors),
	}
	return iterator
}

func (r *Relations) MonitorActors() *RelationIterator {
	r.RLock()
	defer r.RUnlock()

	monitorActors := make([]p.InternalPID, 0, len(r.monitorActors))
	for _, pid := range r.monitorActors {
		monitorActors = append(monitorActors, pid)
	}

	iterator := &RelationIterator{
		data:   monitorActors,
		pos:    0,
		length: len(monitorActors),
	}
	return iterator
}
