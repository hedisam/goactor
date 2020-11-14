package childstate

import "github.com/hedisam/goactor/internal/intlpid"

type ChildrenManager struct {
	children map[string]*ChildState
	// index keeps track of alive children by their internal internal_pid
	index map[intlpid.InternalPID]string
}

func NewChildrenManager() *ChildrenManager {
	return &ChildrenManager{
		children: make(map[string]*ChildState),
		index:    make(map[intlpid.InternalPID]string),
	}
}

func (manager *ChildrenManager) Iterator() *ChildrenStateIterator {
	states := make([]*ChildState, 0, len(manager.children))
	for _, childState := range manager.children {
		states = append(states, childState)
	}
	return newChildrenStateIterator(states)
}

func (manager *ChildrenManager) Index(pid intlpid.InternalPID, name string) {
	manager.index[pid] = name
}

func (manager *ChildrenManager) SearchIndex(pid intlpid.InternalPID) (string, bool) {
	name, ok := manager.index[pid]
	return name, ok
}

func (manager *ChildrenManager) RemoveIndex(pid intlpid.InternalPID) {
	delete(manager.index, pid)
}

func (manager *ChildrenManager) Put(name string, state *ChildState) {
	manager.children[name] = state
}

func (manager *ChildrenManager) Get(name string) (*ChildState, bool) {
	state, ok := manager.children[name]
	return state, ok
}

func (manager *ChildrenManager) GetByPID(pid intlpid.InternalPID) (*ChildState, bool) {
	name, ok := manager.SearchIndex(pid)
	if !ok {
		return nil, false
	}
	state, ok := manager.Get(name)
	return state, ok
}

func (manager *ChildrenManager) Delete(name string) {
	delete(manager.children, name)
}
