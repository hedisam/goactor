package intprocess

import (
	"context"
	"fmt"

	clusteringv1 "github.com/hedisam/goactor/gen/clustering/v1"
)

var _ PID = &NodeProcess{}

type NodeProcess struct {
	ref        string
	dispatcher Dispatcher
}

func NewNodeProcess(ref string, dispatcher Dispatcher) *NodeProcess {
	return &NodeProcess{
		ref:        ref,
		dispatcher: dispatcher,
	}
}

func (p *NodeProcess) Ref() string {
	return p.ref
}

func (p *NodeProcess) PushMessage(ctx context.Context, msg any) error {
	return p.dispatcher.PushMessage(ctx, msg)
}

func (p *NodeProcess) PushSystemMessage(ctx context.Context, msg any) error {
	return p.dispatcher.PushSystemMessage(ctx, msg)
}

func (p *NodeProcess) AcceptLink(linker PID) error {
	err := p.dispatcher.PushSystemMessage(context.Background(), &clusteringv1.SystemRequest{
		Request: &clusteringv1.SystemRequest_Link{
			Link: &clusteringv1.Link{
				LinkerRef: linker.Ref(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("send accept link request to target node actor: %w", err)
	}

	return nil
}

func (p *NodeProcess) AcceptUnlink(linkerRef string) {
	_ = p.dispatcher.PushSystemMessage(context.Background(), &clusteringv1.SystemRequest{
		Request: &clusteringv1.SystemRequest_Unlink{
			Unlink: &clusteringv1.Unlink{
				LinkerRef: linkerRef,
			},
		},
	})
	//if err != nil {
	//	return fmt.Errorf("send accept unlink request to target node actor: %w", err)
	//}
	//
	//return nil
}

func (p *NodeProcess) AcceptMonitor(monitor PID) error {
	err := p.dispatcher.PushSystemMessage(context.Background(), &clusteringv1.SystemRequest{
		Request: &clusteringv1.SystemRequest_Monitor{
			Monitor: &clusteringv1.Monitor{
				MonitorRef: monitor.Ref(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("send accept monitor request to target node actor: %w", err)
	}

	return nil
}

func (p *NodeProcess) AcceptDemonitor(monitorRef string) {
	_ = p.dispatcher.PushSystemMessage(context.Background(), &clusteringv1.SystemRequest{
		Request: &clusteringv1.SystemRequest_Demonitor{
			Demonitor: &clusteringv1.Demonitor{
				MonitorRef: monitorRef,
			},
		},
	})
	//if err != nil {
	//	return fmt.Errorf("send accept demonitor request to target node actor: %w", err)
	//}
	//
	//return nil
}
