package goactor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hedisam/goactor/sysmsg"
	"log/slog"
	"net"
	"reflect"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	clusteringv1 "github.com/hedisam/goactor/internal/gen/clustering/v1"
	"github.com/hedisam/goactor/internal/intprocess"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/registry"
)

var _ clusteringv1.NodeServiceServer = &localNodeServer{}

type localNodeServer struct {
	addr string

	registeredActorsMu   sync.RWMutex
	sigToRegisteredActor map[string]ActorFactory

	remoteRefToPID map[string]intprocess.PID
	remoteRefMu    sync.RWMutex
	remoteNodeReg  *registry.RemoteNodeRegistry
}

func startLocalNodeServer() (*localNodeServer, error) {
	l, err := net.Listen("tcp", "localhost:")
	if err != nil {
		return nil, fmt.Errorf("start tcp listener on localhost: %w", err)
	}

	ns := &localNodeServer{
		addr:                 l.Addr().String(),
		sigToRegisteredActor: make(map[string]ActorFactory),
		remoteRefToPID:       make(map[string]intprocess.PID),
		remoteNodeReg:        registry.NewRemoteNodeRegistry(logger),
	}

	s := grpc.NewServer()
	healthServer := health.NewServer()
	healthgrpc.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", healthgrpc.HealthCheckResponse_SERVING)
	clusteringv1.RegisterNodeServiceServer(s, ns)
	go func() {
		logger.Debug("Starting node server", "addr", l.Addr().String())
		err = s.Serve(l)
		if err != nil {
			logger.Error("Clustering grpc server failed with error", slog.Any("error", err))
		}
	}()

	return ns, nil
}

func (s *localNodeServer) Spawn(ctx context.Context, req *clusteringv1.SpawnRequest) (*clusteringv1.SpawnResponse, error) {
	actorFactory, ok := s.getRegisteredActorFactory(req.GetActorSignature())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such actor registered for node spawn: %q", req.GetActorSignature())
	}

	actor := actorFactory()
	err := json.Unmarshal(req.GetActorData(), &actor)
	if err != nil {
		return nil, fmt.Errorf("unmarshal actor data: %w", err)
	}

	pid, err := Spawn(context.WithoutCancel(ctx), actor)
	if err != nil {
		return nil, fmt.Errorf("could not spawn actor: %w", err)
	}

	return &clusteringv1.SpawnResponse{
		Ref: pid.Ref(),
	}, nil
}

func (s *localNodeServer) Send(ctx context.Context, req *clusteringv1.SendRequest) (*clusteringv1.SendResponse, error) {
	pid, ok := registry.LocalProcessByRef(req.GetRecipientRef())
	if !ok {
		return nil, fmt.Errorf("no running actor available with the given ID %q", req.GetRecipientRef())
	}

	if sysMessage := req.GetSystemMessage(); sysMessage != nil {
		err := s.handleSystemMessage(pid, sysMessage)
		if err != nil {
			return nil, fmt.Errorf("handle system message: %w", err)
		}
		return &clusteringv1.SendResponse{}, nil
	}

	if internalMessage := req.GetInternalMessage(); internalMessage != nil {
		var m map[string]*sysmsg.Message
		err := json.Unmarshal(internalMessage.GetData(), &m)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal internal message: %w", err)
		}
		err = pid.PushSystemMessage(ctx, m["data"])
		if err != nil {
			return nil, fmt.Errorf("could not push internal message: %w", err)
		}
		return &clusteringv1.SendResponse{}, nil
	}

	message := req.GetUserMessage()
	if message == nil {
		return nil, fmt.Errorf("unknown message type: %T", req.GetMessage())
	}

	msg, err := unmarshalNodeMessage(message.GetData())
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal message: %w", err)
	}

	err = Send(ctx, &PID{internalPID: pid}, msg)
	if err != nil {
		return nil, fmt.Errorf("send to pid: %w", err)
	}

	return &clusteringv1.SendResponse{}, nil
}

func (s *localNodeServer) handleSystemMessage(pid *intprocess.LocalProcess, msg *clusteringv1.SystemMessage) error {
	senderPID, err := s.getOrCreateRemotePID(msg.GetSenderRef(), msg.GetSenderNode())
	if err != nil {
		return fmt.Errorf("could not get system message pid: %w", err)
	}

	switch msg.GetRequest().(type) {
	case *clusteringv1.SystemMessage_Link:
		err = senderPID.Link(pid)
		if err != nil {
			return fmt.Errorf("link: %w", err)
		}
		err = pid.AcceptLink(senderPID)
		if err != nil {
			return fmt.Errorf("linkee accept link: %w", err)
		}
		return nil
	case *clusteringv1.SystemMessage_Unlink:
		err = senderPID.Unlink(pid)
		if err != nil {
			return fmt.Errorf("unlink: %w", err)
		}
		pid.AcceptUnlink(senderPID.Ref())
		return nil
	case *clusteringv1.SystemMessage_Monitor:
		err = senderPID.Demonitor(pid)
		if err != nil {
			return fmt.Errorf("monitor: %w", err)
		}
		err = pid.AcceptMonitor(senderPID)
		if err != nil {
			return fmt.Errorf("monitoree accept monitor: %w", err)
		}
		return nil
	case *clusteringv1.SystemMessage_Demonitor:
		err = senderPID.Demonitor(pid)
		if err != nil {
			return fmt.Errorf("demonitor: %w", err)
		}
		pid.AcceptDemonitor(senderPID.Ref())
		return nil
	default:
		return fmt.Errorf("unknown node system request received: %T", msg.GetRequest())
	}
}

func (s *localNodeServer) getRegisteredActorFactory(sig string) (ActorFactory, bool) {
	s.registeredActorsMu.RLock()
	defer s.registeredActorsMu.RUnlock()
	actor, ok := s.sigToRegisteredActor[sig]
	return actor, ok
}

func (s *localNodeServer) registerActorType(actor Actor, factory ActorFactory) {
	sig := generateActorTypeSig(actor)
	s.registeredActorsMu.Lock()
	s.sigToRegisteredActor[sig] = factory
	s.registeredActorsMu.Unlock()
}

func (s *localNodeServer) unregisterActorType(actor Actor) {
	sig := generateActorTypeSig(actor)
	s.registeredActorsMu.Lock()
	delete(s.sigToRegisteredActor, sig)
	s.registeredActorsMu.Unlock()
}

func (s *localNodeServer) getOrCreateRemotePID(ref, nodeAddr string) (intprocess.PID, error) {
	s.remoteRefMu.RLock()
	remotePID, ok := s.remoteRefToPID[ref]
	if ok {
		s.remoteRefMu.RUnlock()
		return remotePID, nil
	}
	s.remoteRefMu.RUnlock()

	client, err := s.remoteNodeReg.GetOrCreateClient(nodeAddr)
	if err != nil {
		return nil, fmt.Errorf("could not get or create sender's conn: %w", err)
	}

	pid := intprocess.NewNodeProcess(
		logger,
		client.CloseCh,
		ref,
		nodeAddr,
		mailbox.NewGRPCDispatcher(
			client.Client,
			marshalNodeMessage,
			ref,
		),
	)
	s.remoteRefMu.Lock()
	s.remoteRefToPID[ref] = pid
	s.remoteRefMu.Unlock()

	return pid, nil
}

func marshalNodeMessage(msg any) ([]byte, error) {
	data, err := json.Marshal(map[string]any{
		"data": msg,
	})
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}
	return data, nil
}

func unmarshalNodeMessage(data []byte) (any, error) {
	var m map[string]any
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	return m["data"], nil
}

func generateActorTypeSig(actor Actor) string {
	typ := reflect.TypeOf(actor)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return fmt.Sprintf("%s/%s", typ.PkgPath(), typ.Name())
}
