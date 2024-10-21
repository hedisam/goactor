package intprocess

import "github.com/hedisam/goactor/internal/mailbox"

type NodeProcess struct {
	ref           string
	msgDispatcher mailbox.Dispatcher
}

func NewNodeProcess(ref string, dispatcher mailbox.Dispatcher) *NodeProcess {
	return &NodeProcess{
		ref:           ref,
		msgDispatcher: dispatcher,
	}
}

func (p *NodeProcess) Ref() string {
	return p.ref
}

func (p *NodeProcess) Link(pid PID) error {
	//TODO implement me
	panic("implement me")
}

func (p *NodeProcess) Unlink(pid PID) error {
	//TODO implement me
	panic("implement me")
}

func (p *NodeProcess) Monitor(pid PID) error {
	//TODO implement me
	panic("implement me")
}

func (p *NodeProcess) Demonitor(pid PID) error {
	//TODO implement me
	panic("implement me")
}

func (p *NodeProcess) SetTrapExit(trapExit bool) {}

func (p *NodeProcess) Dispatcher() mailbox.Dispatcher {
	return p.msgDispatcher
}

func (p *NodeProcess) disposed() bool {
	//TODO implement me
	panic("implement me")
}

func (p *NodeProcess) addRelation(pid PID, typ relationType) {
	//TODO implement me
	panic("implement me")
}

func (p *NodeProcess) remRelation(ref string, typ relationType) {
	//TODO implement me
	panic("implement me")
}
