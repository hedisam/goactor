package intprocess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"

	clusteringv1 "github.com/hedisam/goactor/internal/gen/clustering/v1"
	"github.com/hedisam/goactor/sysmsg"
)

var _ PID = &NodeProcess{}

type NodeProcess struct {
	logger        *slog.Logger
	ref           string
	localNodeAddr string
	dispatcher    dispatcher
	relations     *relations
	disposedFlag  atomic.Bool
}

func NewNodeProcess(logger *slog.Logger, connectionCloseChan <-chan struct{}, ref, localNodeAddr string, dispatcher dispatcher) *NodeProcess {
	p := &NodeProcess{
		logger:        logger,
		ref:           ref,
		localNodeAddr: localNodeAddr,
		dispatcher:    dispatcher,
		relations:     newRelations(),
	}
	p.monitorConnection(connectionCloseChan)
	return p
}

func (p *NodeProcess) Ref() string {
	return p.ref
}

func (p *NodeProcess) SendMessage(ctx context.Context, msg any) error {
	return p.dispatcher.PushMessage(ctx, msg)
}

func (p *NodeProcess) SendSystemMessage(ctx context.Context, msg any) error {
	return p.dispatcher.PushSystemMessage(ctx, msg)
}

func (p *NodeProcess) Link(linkee PID) error {
	if p.disposed() {
		return ErrSelfDisposed
	}

	p.relations.Add(linkee, relationLinked)
	return nil
}

func (p *NodeProcess) Unlink(linkee PID) error {
	if p.disposed() {
		return ErrSelfDisposed
	}
	p.relations.Remove(linkee.Ref(), relationLinked)
	return nil
}

func (p *NodeProcess) Monitor(monitored PID) error {
	if p.disposed() {
		return ErrSelfDisposed
	}
	p.relations.Add(monitored, relationMonitored)
	return nil
}

func (p *NodeProcess) Demonitor(monitored PID) error {
	if p.disposed() {
		return ErrSelfDisposed
	}
	p.relations.Remove(monitored.Ref(), relationMonitored)
	return nil
}

func (p *NodeProcess) AcceptLink(linker PID) error {
	if p.disposed() {
		return ErrTargetDisposed
	}

	err := p.dispatcher.PushSystemMessage(context.Background(), &clusteringv1.SystemMessage{
		SenderRef:  linker.Ref(),
		SenderNode: p.localNodeAddr,
		Request: &clusteringv1.SystemMessage_Link{
			Link: &clusteringv1.Link{},
		},
	})
	if err != nil {
		return fmt.Errorf("send accept link request to target node actor: %w", err)
	}

	p.relations.Add(linker, relationLinked)

	return nil
}

func (p *NodeProcess) AcceptUnlink(linkerRef string) {
	if p.disposed() {
		return
	}

	p.relations.Remove(linkerRef, relationLinked)

	_ = p.dispatcher.PushSystemMessage(context.Background(), &clusteringv1.SystemMessage{
		SenderRef:  linkerRef,
		SenderNode: p.localNodeAddr,
		Request: &clusteringv1.SystemMessage_Unlink{
			Unlink: &clusteringv1.Unlink{},
		},
	})
}

func (p *NodeProcess) AcceptMonitor(monitor PID) error {
	if p.disposed() {
		return ErrTargetDisposed
	}

	err := p.dispatcher.PushSystemMessage(context.Background(), &clusteringv1.SystemMessage{
		SenderRef:  monitor.Ref(),
		SenderNode: p.localNodeAddr,
		Request: &clusteringv1.SystemMessage_Monitor{
			Monitor: &clusteringv1.Monitor{},
		},
	})
	if err != nil {
		return fmt.Errorf("send accept monitor request to target node actor: %w", err)
	}

	p.relations.Add(monitor, relationMonitor)

	return nil
}

func (p *NodeProcess) AcceptDemonitor(monitorRef string) {
	if p.disposed() {
		return
	}

	p.relations.Remove(monitorRef, relationMonitor)

	_ = p.dispatcher.PushSystemMessage(context.Background(), &clusteringv1.SystemMessage{
		Request: &clusteringv1.SystemMessage_Demonitor{
			Demonitor: &clusteringv1.Demonitor{},
		},
	})
}

func (p *NodeProcess) disposed() bool {
	return p == nil || p.disposedFlag.Load()
}

func (p *NodeProcess) monitorConnection(closeCh <-chan struct{}) {
	go func() {
		// todo: this only monitors the client connection; we should also terminate if the remote actor is exited.
		<-closeCh
		p.disposedFlag.Store(true)
		relationTypeToPIDs := p.relations.TypeToRelatedPIDs()

		reason := errors.New("no connection")
		p.logger.Info("Node actor is getting disposed due to connection error, notifying related actors",
			slog.String("actor", p.ref),
			slog.String("reason", reason.Error()),
		)

		notify(context.Background(), p.logger, p.ref, sysmsg.Exit, reason, relationTypeToPIDs[relationLinked]...)
		notify(context.Background(), p.logger, p.ref, sysmsg.Down, reason, relationTypeToPIDs[relationMonitor]...)
	}()
}
