package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"time"
)

type supRef struct {
	pid *p.PID
}

func ToSupervisorRef(pid *p.PID) (*supRef, error) {
	if pid.IsSupervisor() {
		return &supRef{pid: pid}, nil
	}
	return nil, fmt.Errorf("can not convert a worker pid to supervisor reference")
}

func (ref *supRef) ChildrenCount(timeout time.Duration) (*ChildrenCount, error) {
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

func (ref *supRef) DeleteChild(name string, timeout time.Duration) error {
	req := NewDeleteChildRequest(name)
	_, err := ref.request(req, timeout)
	return err
}

func (ref *supRef) RestartChild(name string, timeout time.Duration) error {
	req := NewRestartChildRequest(name)
	_, err := ref.request(req, timeout)
	return err
}

func (ref *supRef) StartNewChild(spec Spec, timeout time.Duration) error {
	req := NewStartChildRequest(spec)
	_, err := ref.request(req, timeout)
	return err
}

func (ref *supRef) TerminateChild(name string, timeout time.Duration) error {
	req := NewTerminateChildRequest(name)
	_, err := ref.request(req, timeout)
	return err
}

func (ref *supRef) request(request refRequest, timeout time.Duration) (resp interface{}, err error) {
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
