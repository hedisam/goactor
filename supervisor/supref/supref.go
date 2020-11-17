package supref

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/spec"
	"time"
)

type SupRef struct {
	pid *p.PID
}

func ToSupervisorRef(pid *p.PID) (*SupRef, error) {
	if pid.IsSupervisor() {
		return &SupRef{pid: pid}, nil
	}
	return nil, fmt.Errorf("can not convert a worker pid to supervisor reference")
}

// PID returns the supervisor's pid. Note that it can be used by users to send message to the supervisor.
// To interact with the supervisor, you must use the supervisor's reference SupRef.
func (ref *SupRef) PID() *p.PID {
	return ref.pid
}

// ChildrenCount returns a *ChildrenCount object which denotes the count of all type of supervisor's children.
func (ref *SupRef) ChildrenCount(timeout time.Duration) (*ChildrenCount, error) {
	req := NewChildrenCountRequest()
	resp, err := ref.request(req, timeout)
	if err != nil {
		return nil, err
	}
	response, ok := resp.(*ChildrenCount)
	if !ok {
		return nil, fmt.Errorf("unknown response from supervisor: %v", resp)
	}
	return response, nil
}

func (ref *SupRef) DeleteChild(name string, timeout time.Duration) error {
	req := NewDeleteChildRequest(name)
	_, err := ref.request(req, timeout)
	return err
}

func (ref *SupRef) RestartChild(name string, timeout time.Duration) error {
	req := NewRestartChildRequest(name)
	_, err := ref.request(req, timeout)
	return err
}

func (ref *SupRef) StartNewChild(spec spec.Spec, timeout time.Duration) error {
	req := NewStartChildRequest(spec)
	_, err := ref.request(req, timeout)
	return err
}

func (ref *SupRef) TerminateChild(name string, timeout time.Duration) error {
	req := NewTerminateChildRequest(name)
	_, err := ref.request(req, timeout)
	return err
}

func (ref *SupRef) request(request refRequest, timeout time.Duration) (resp interface{}, err error) {
	future := goactor.NewFutureActor()

	request.SetRequester(future.Self().InternalPID())

	err = intlpid.SendSystemMessage(ref.pid.InternalPID(), request)
	if err != nil {
		return nil, fmt.Errorf("couldn't deliver request to the supervisor: %w'", err)
	}
	// waiting for a reply
	future.ReceiveWithTimeout(timeout, func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case error:
			err = msg
		case supRefResponse:
			resp = msg
		default:
			err = fmt.Errorf("unknown response from the supervisor: %v", message)
		}
		return false
	})
	return
}
