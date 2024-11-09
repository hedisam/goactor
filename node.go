package goactor

import (
	"context"
	"encoding/json"
	"fmt"

	clusteringv1 "github.com/hedisam/goactor/internal/gen/clustering/v1"
	"github.com/hedisam/goactor/internal/intprocess"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/registry"
)

var node *LocalNode

type LocalNode struct {
	server   *localNodeServer
	nodesReg *registry.RemoteNodeRegistry
}

func initLocalNode(server *localNodeServer) {
	node = &LocalNode{
		server:   server,
		nodesReg: registry.NewRemoteNodeRegistry(logger),
	}
}

func Node() *LocalNode {
	return node
}

// Spawn spawns a remote actor and registers it with the local node actor.
func (n *LocalNode) Spawn(ctx context.Context, nodeAddr string, actor Actor) (*PID, error) {
	client, err := n.nodesReg.GetOrCreateClient(nodeAddr)
	if err != nil {
		return nil, fmt.Errorf("could not get or create client node: %w", err)
	}

	data, err := json.Marshal(actor)
	if err != nil {
		return nil, fmt.Errorf("json marshal actor data: %w", err)
	}

	actorSig := generateActorTypeSig(actor)
	resp, err := client.Client.Spawn(ctx, &clusteringv1.SpawnRequest{
		ActorSignature: actorSig,
		ActorData:      data,
	})
	if err != nil {
		return nil, fmt.Errorf("spawn remote actor: %w", err)
	}

	pid := &PID{
		internalPID: intprocess.NewNodeProcess(
			logger,
			client.CloseCh,
			resp.GetRef(),
			n.Addr(),
			mailbox.NewGRPCDispatcher(
				client.Client,
				marshalNodeMessage,
				resp.GetRef(),
			),
		),
	}

	return pid, nil
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
