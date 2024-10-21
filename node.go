package goactor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	clusteringv1 "github.com/hedisam/goactor/gen/clustering/v1"
	"github.com/hedisam/goactor/internal/intprocess"
	"github.com/hedisam/goactor/internal/mailbox"
)

var node *LocalNode

type LocalNode struct {
	server *localNodeServer

	nodesGroup     singleflight.Group
	nodeConnsMu    sync.RWMutex
	nodeAddrToConn map[string]*grpcConn
}

func initLocalNode(server *localNodeServer) {
	node = &LocalNode{
		server:         server,
		nodeAddrToConn: make(map[string]*grpcConn),
	}
}

func Node() *LocalNode {
	return node
}

func (n *LocalNode) Spawn(ctx context.Context, nodeAddr string, actor Actor) (*PID, error) {
	conn, err := n.getNodeConn(nodeAddr)
	if err != nil {
		return nil, fmt.Errorf("could not get or create client node: %w", err)
	}

	data, err := json.Marshal(actor)
	if err != nil {
		return nil, fmt.Errorf("json marshal actor data: %w", err)
	}

	actorSig := generateActorTypeSig(actor)
	clusteringClient := clusteringv1.NewNodeServiceClient(conn.conn)
	resp, err := clusteringClient.Spawn(ctx, &clusteringv1.SpawnRequest{
		ActorSignature: actorSig,
		ActorData:      data,
	})
	if err != nil {
		return nil, fmt.Errorf("spawn remote actor: %w", err)
	}

	return &PID{
		internalPID: intprocess.NewNodeProcess(
			resp.GetRef(),
			mailbox.NewGRPCDispatcher(
				clusteringClient,
				marshalNodeMessage,
				resp.GetRef(),
			),
		),
	}, nil
}

func (n *LocalNode) RegisterActorType(actor Actor, factory ActorFactory) {
	n.server.registerActorType(actor, factory)
}

func (n *LocalNode) UnregisterActorType(actor Actor) {
	n.server.unregisterActorType(actor)
}

func (n *LocalNode) Addr() string {
	return n.server.addr
}

func (n *LocalNode) getNodeConn(addr string) (*grpcConn, error) {
	var conn *grpcConn
	_, err, _ := n.nodesGroup.Do(addr, func() (any, error) {
		var err error
		conn, err = n._getOrCreateNodeConn(addr)
		if err != nil {
			return nil, fmt.Errorf(" get or create node conn: %w", err)
		}
		return conn, nil
	})
	return conn, err
}

// _getOrCreateNodeConn only to be called from getNodeConn.
func (n *LocalNode) _getOrCreateNodeConn(addr string) (*grpcConn, error) {
	n.nodeConnsMu.RLock()
	c, ok := n.nodeAddrToConn[addr]
	n.nodeConnsMu.RUnlock()
	if ok {
		return c, nil
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{
				"healthCheckConfig": {
					"serviceName": ""
				}
			}`),
	)
	if err != nil {
		return nil, fmt.Errorf("create new grpc client connection: %w", err)
	}

	c = &grpcConn{conn: conn}
	n.nodeConnsMu.Lock()
	n.nodeAddrToConn[addr] = c
	n.nodeConnsMu.Unlock()
	return c, nil
}
