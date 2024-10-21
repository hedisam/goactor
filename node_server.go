package goactor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"reflect"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	clusteringv1 "github.com/hedisam/goactor/gen/clustering/v1"
	"github.com/hedisam/goactor/internal/registry"
)

var _ clusteringv1.NodeServiceServer = &localNodeServer{}

type grpcConn struct {
	conn *grpc.ClientConn
	mu   sync.RWMutex
}

type localNodeServer struct {
	addr string

	registeredActorsMu   sync.RWMutex
	sigToRegisteredActor map[string]ActorFactory
}

func startLocalNodeServer() (*localNodeServer, error) {
	l, err := net.Listen("tcp", "localhost:")
	if err != nil {
		return nil, fmt.Errorf("start tcp listener on localhost: %w", err)
	}

	ns := &localNodeServer{
		addr:                 l.Addr().String(),
		sigToRegisteredActor: make(map[string]ActorFactory),
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
	pid, ok := registry.GetRegistry().PIDByRef(req.GetRef())
	if !ok {
		return nil, fmt.Errorf("no running actor available with the given ID %q", req.GetRef())
	}

	msg, err := unmarshalNodeMessage(req.GetMessage().GetData())
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal message: %w", err)
	}

	err = Send(ctx, &PID{internalPID: pid}, msg)
	if err != nil {
		return nil, fmt.Errorf("send to pid: %w", err)
	}

	return &clusteringv1.SendResponse{}, nil
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
