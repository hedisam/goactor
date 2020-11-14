package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type refRequest interface {
	SetRequester(pid intlpid.InternalPID)
}

type refBaseRequest struct {
	service *SupService
	requester intlpid.InternalPID
}

func (req *refBaseRequest) SetRequester(pid intlpid.InternalPID) {
	req.requester = pid
}

func (req *refBaseRequest) Reply(tag string, resp interface{}) {
	err := intlpid.SendMessage(req.requester, resp)
	if err != nil {
		log.Printf("[!] supervisor couldn't send response to a %s - err: %v\n", tag, err)
	}
}

// SetSupervisorService must be called before calling the Run method
func (req *refBaseRequest) SetSupervisorService(service *SupService) {
	req.service = service
}

type ChildrenCountRequest struct {
	*refBaseRequest
}

func NewChildrenCountRequest() *ChildrenCountRequest {
	return &ChildrenCountRequest{&refBaseRequest{}}
}

func (req *ChildrenCountRequest) Run(_ sysmsg.SystemMessage) bool {
	childrenIterator := req.service.ChildrenIterator()
	resp := &ChildrenCount{}

	// all children count, dead or alive
	resp.Specs = childrenIterator.Size()

	for childrenIterator.HasNext() {
		child := childrenIterator.Value()

		if !child.Dead() {
			resp.Active++
		}
		if child.IsSupervisor() {
			resp.Supervisors++
			continue
		}
		resp.Workers++
	}

	req.Reply("ChildrenCountRequest", resp)

	return true
}

type DeleteChildRequest struct {
	*refBaseRequest
	name string
}

func NewDeleteChildRequest(name string) *DeleteChildRequest {
	return &DeleteChildRequest{
		refBaseRequest: &refBaseRequest{},
		name:           name,
	}
}

func (req *DeleteChildRequest) Run(_ sysmsg.SystemMessage) bool {
	tag := "DeleteChildRequest"

	child, ok := req.service.GetChildByName(req.name)
	if !ok {
		req.Reply(tag, fmt.Errorf("child not found with the given name: %s", req.name))
		return true
	}

	if !child.Dead() {
		req.Reply(tag, fmt.Errorf("child '%s' is alive, you have to first terminate it", req.name))
		return true
	}

	err := req.service.DeleteChild(child)
	if err != nil {
		req.Reply(tag, fmt.Errorf("couldn't delete the child: %w", err))
		return true
	}

	// child deleted successfully
	req.Reply(tag, &OK{})

	return true
}

type RestartChildRequest struct {
	*refBaseRequest
	name string
}

func NewRestartChildRequest(name string) *RestartChildRequest {
	return &RestartChildRequest{
		refBaseRequest: &refBaseRequest{},
		name:           name,
	}
}

func (req *RestartChildRequest) Run(_ sysmsg.SystemMessage) bool {
	tag := "RestartChildRequest"

	child, ok := req.service.GetChildByName(req.name)
	if !ok {
		req.Reply(tag, fmt.Errorf("child not found with the given name: %s", req.name))
		return true
	}

	if !child.Dead() {
		req.Reply(tag, fmt.Errorf("can not restart a running child '%s'", req.name))
		return true
	}

	err := child.Restart()
	if err != nil {
		req.Reply(tag, fmt.Errorf("failed restarting child process: %w", err))
		return true
	}

	// child restarted successfully
	req.Reply(tag, &OK{})

	return true
}

type StartChildRequest struct {
	*refBaseRequest
	spec Spec
}

func NewStartChildRequest(spec Spec) *StartChildRequest {
	return &StartChildRequest{
		refBaseRequest: &refBaseRequest{},
		spec:           spec,
	}
}

func (req *StartChildRequest) Run(_ sysmsg.SystemMessage) bool {
	tag := "StartChildRequest"

	err := validateSpec(req.spec)
	if err != nil {
		req.Reply(tag, fmt.Errorf("invalid child spec: %w", err))
		return true
	}

	err = req.service.StartChild(req.spec)
	if err != nil {
		req.Reply(tag, fmt.Errorf("failed spawning the new child: %w", err))
		return true
	}

	req.Reply(tag, &OK{})
	return true
}

type TerminateChildRequest struct {
	*refBaseRequest
	name string
}

func NewTerminateChildRequest(name string) *TerminateChildRequest {
	return &TerminateChildRequest{
		refBaseRequest: &refBaseRequest{},
		name:           name,
	}
}

func (req *TerminateChildRequest) Run(_ sysmsg.SystemMessage) bool {
	tag := "TerminateChildRequest"

	child, ok := req.service.GetChildByName(req.name)
	if !ok {
		req.Reply(tag, fmt.Errorf("child not found with the given name: %s", req.name))
		return true
	}

	if child.Dead() {
		req.Reply(tag, fmt.Errorf("child '%s' already has been terminated", req.name))
		return true
	}

	err := req.service.shutdownChild(
		child,
		sysmsg.NewKillMessage(req.service.Self().InternalPID(), "terminated by user's request", nil),
		)
	if err != nil {
		req.Reply(tag, fmt.Errorf("failed to terminate the child: %w", err))
		return true
	}

	req.Reply(tag, &OK{})
	return true
}







