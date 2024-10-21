package mailbox

import (
	"context"
	"fmt"

	clusteringv1 "github.com/hedisam/goactor/gen/clustering/v1"
)

type MessageMarshaller func(msg any) ([]byte, error)

// GRPCDispatcher acts only as a dispatcher and will not support receiving messages here.
type GRPCDispatcher struct {
	client     clusteringv1.NodeServiceClient
	marshaller MessageMarshaller
	selfRef    string
}

func NewGRPCDispatcher(client clusteringv1.NodeServiceClient, marshaller MessageMarshaller, ref string) *GRPCDispatcher {
	return &GRPCDispatcher{
		client:     client,
		marshaller: marshaller,
		selfRef:    ref,
	}
}

func (m *GRPCDispatcher) PushMessage(ctx context.Context, msg any) error {
	return m.push(ctx, msg, false)
}

func (m *GRPCDispatcher) PushSystemMessage(ctx context.Context, msg any) error {
	return m.push(ctx, msg, true)
}

func (m *GRPCDispatcher) push(ctx context.Context, msg any, sysMessage bool) error {
	data, err := m.marshaller(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	_, err = m.client.Send(ctx, &clusteringv1.SendRequest{
		Ref: m.selfRef,
		Message: &clusteringv1.Message{
			Data:        data,
			IsSystemMsg: sysMessage,
		},
	})
	if err != nil {
		return fmt.Errorf("send via node client: %w", err)
	}
	return nil
}
