package registry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	clusteringv1 "github.com/hedisam/goactor/internal/gen/clustering/v1"
)

type Client struct {
	Client  clusteringv1.NodeServiceClient
	CloseCh <-chan struct{}
}

type RemoteNodeRegistry struct {
	logger       *slog.Logger
	g            singleflight.Group
	mu           sync.RWMutex
	addrToClient map[string]*Client
}

func NewRemoteNodeRegistry(logger *slog.Logger) *RemoteNodeRegistry {
	return &RemoteNodeRegistry{
		logger:       logger,
		addrToClient: make(map[string]*Client),
	}
}

func (r *RemoteNodeRegistry) GetOrCreateClient(addr string) (*Client, error) {
	clientAny, err, _ := r.g.Do(addr, func() (any, error) {
		client, err := r.getOrCreateNodeClient(addr)
		if err != nil {
			return nil, fmt.Errorf("get or create node client: %w", err)
		}
		return client, nil
	})
	if err != nil {
		return nil, err
	}

	client, _ := clientAny.(*Client)
	return client, nil
}

func (r *RemoteNodeRegistry) getOrCreateNodeClient(addr string) (*Client, error) {
	r.mu.RLock()
	c, ok := r.addrToClient[addr]
	r.mu.RUnlock()
	if ok {
		return c, nil
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("create new grpc client connection: %w", err)
	}

	client, err := r.registerAndMonitor(addr, conn)
	if err != nil {
		return nil, fmt.Errorf("register and monitor node client: %w", err)
	}

	return client, nil
}

func (r *RemoteNodeRegistry) registerAndMonitor(addr string, conn *grpc.ClientConn) (*Client, error) {
	healthClient := healthgrpc.NewHealthClient(conn)
	stream, err := healthClient.Watch(context.Background(), &healthgrpc.HealthCheckRequest{})
	if err != nil {
		return nil, fmt.Errorf("could not watch remote node: %w", err)
	}

	closeCh := make(chan struct{})
	client := &Client{
		Client:  clusteringv1.NewNodeServiceClient(conn),
		CloseCh: closeCh,
	}
	r.mu.Lock()
	r.addrToClient[addr] = client
	r.mu.Unlock()

	go func() {
		defer func() {
			r.mu.Lock()
			delete(r.addrToClient, addr)
			r.mu.Unlock()

			close(closeCh)
		}()

		for {
			_, err := stream.Recv()
			if err != nil {
				return
			}
		}
	}()

	return client, nil
}
