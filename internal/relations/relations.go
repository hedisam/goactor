package relations

import (
	p "github.com/hedisam/goactor/internal/intlpid"
	"sync"
)

type RelationType int32
type RelationMap map[p.InternalPID]struct{}

const (
	LinkedRelation RelationType = iota
	MonitoredRelation
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
	if _, ok := r.linkedActors[pid]; ok {
		return LinkedRelation
	} else if _, ok := r.monitoredActors[pid]; ok {
		return MonitoredRelation
	}
	return NoRelation
}

func (r *Relations) AddLink(to p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	r.linkedActors[to] = struct{}{}
}

func (r *Relations) RemoveLink(from p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	delete(r.linkedActors, from)
}

func (r *Relations) AddMonitored(who p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	r.monitoredActors[who] = struct{}{}
}

func (r *Relations) RemoveMonitored(who p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	delete(r.monitoredActors, who)
}

func (r *Relations) AddMonitor(by p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	r.monitorActors[by] = struct{}{}
}

func (r *Relations) RemoveMonitor(by p.InternalPID) {
	r.Lock()
	defer r.Unlock()
	delete(r.monitorActors, by)
}

func (r *Relations) LinkedActors() *RelationIterator {
	r.RLock()
	defer r.RUnlock()

	linkedActors := make([]p.InternalPID, 0, len(r.linkedActors))
	for pid := range r.linkedActors {
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
	for pid := range r.monitorActors {
		monitorActors = append(monitorActors, pid)
	}

	iterator := &RelationIterator{
		data:   monitorActors,
		pos:    0,
		length: len(monitorActors),
	}
	return iterator
}
