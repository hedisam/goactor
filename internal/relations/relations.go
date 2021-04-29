package relations

import (
	"fmt"
	p "github.com/hedisam/goactor/internal/intlpid"
	"sync"
	"sync/atomic"
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
	disposed int32
}

func NewRelation() *Relations {
	return &Relations{
		linkedActors:    make(RelationMap),
		monitorActors:   make(RelationMap),
		monitoredActors: make(RelationMap),
	}
}

func (r *Relations) Dispose() {
	atomic.StoreInt32(&r.disposed, 1)
}

func (r *Relations) RelationType(pid p.InternalPID) RelationType {
	if pid == nil {
		return NoRelation
	}

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

func (r *Relations) AddLink(to p.InternalPID) error {
	if to == nil {
		return fmt.Errorf("AddLink failed: nil pid")
	}
	if atomic.LoadInt32(&r.disposed) == 1 {
		return fmt.Errorf("AddLink failed: disposed relation manager")
	}
	r.Lock()
	defer r.Unlock()

	r.linkedActors[to.ID()] = to
	return nil
}

func (r *Relations) RemoveLink(from p.InternalPID) error {
	if from == nil {
		return fmt.Errorf("RemoveLink failed: nil pid")
	}
	if atomic.LoadInt32(&r.disposed) == 1 {
		return fmt.Errorf("RemoveLink failed: disposed relation manager")
	}
	r.Lock()
	defer r.Unlock()

	delete(r.linkedActors, from.ID())
	return nil
}

func (r *Relations) AddMonitored(who p.InternalPID) error {
	if who == nil {
		return fmt.Errorf("AddMonitored failed: nil pid")
	}
	if atomic.LoadInt32(&r.disposed) == 1 {
		return fmt.Errorf("AddMonitored failed: disposed relation manager")
	}
	r.Lock()
	defer r.Unlock()

	r.monitoredActors[who.ID()] = who
	return nil
}

func (r *Relations) RemoveMonitored(who p.InternalPID) error {
	if who == nil {
		return fmt.Errorf("RemoveMonitored failed: nil pid")
	}
	if atomic.LoadInt32(&r.disposed) == 1 {
		return fmt.Errorf("RemoveMonitored failed: disposed relation manager")
	}
	r.Lock()
	defer r.Unlock()

	delete(r.monitoredActors, who.ID())
	return nil
}

func (r *Relations) AddMonitor(by p.InternalPID) error {
	if by == nil {
		return fmt.Errorf("AddMonitor failed: nil pid")
	}
	if atomic.LoadInt32(&r.disposed) == 1 {
		return fmt.Errorf("AddMonitor failed: disposed relation manager")
	}
	r.Lock()
	defer r.Unlock()

	r.monitorActors[by.ID()] = by
	return nil
}

func (r *Relations) RemoveMonitor(by p.InternalPID) error {
	if by == nil {
		return fmt.Errorf("RemoveMonitor failed: nil pid")
	}
	if atomic.LoadInt32(&r.disposed) == 1 {
		return fmt.Errorf("RemoveMonitor failed: disposed relation manager")
	}
	r.Lock()
	defer r.Unlock()

	delete(r.monitorActors, by.ID())
	return nil
}

func (r *Relations) LinkedActors() *RelationIterator {
	r.RLock()
	defer r.RUnlock()

	linkedActors := make([]p.InternalPID, 0, len(r.linkedActors))
	for _, pid := range r.linkedActors {
		linkedActors = append(linkedActors, pid)
	}

	return NewRelationIterator(linkedActors)
}

func (r *Relations) MonitorActors() *RelationIterator {
	r.RLock()
	defer r.RUnlock()

	monitorActors := make([]p.InternalPID, 0, len(r.monitorActors))
	for _, pid := range r.monitorActors {
		monitorActors = append(monitorActors, pid)
	}

	return NewRelationIterator(monitorActors)
}
