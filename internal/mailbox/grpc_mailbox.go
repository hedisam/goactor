package mailbox

import (
	"context"
	"fmt"

	clusteringv1 "github.com/hedisam/goactor/internal/gen/clustering/v1"
	"github.com/hedisam/goactor/sysmsg"
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
	data, err := m.marshaller(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	_, err = m.client.Send(ctx, &clusteringv1.SendRequest{
		RecipientRef: m.selfRef,
		Message: &clusteringv1.SendRequest_UserMessage{
			UserMessage: &clusteringv1.UserMessage{
				Data: data,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("user message via node dispatcher: %w", err)
	}
	return nil
}

func (m *GRPCDispatcher) PushSystemMessage(ctx context.Context, msg any) error {
	req := &clusteringv1.SendRequest{
		RecipientRef: m.selfRef,
	}

	switch t := msg.(type) {
	case *clusteringv1.SystemMessage:
		req.Message = &clusteringv1.SendRequest_SystemMessage{
			SystemMessage: t,
		}
	case *sysmsg.Message:
		data, err := m.marshaller(msg)
		if err != nil {
			return fmt.Errorf("could not marshal internal message: %w", err)
		}
		req.Message = &clusteringv1.SendRequest_InternalMessage{
			InternalMessage: &clusteringv1.InternalMessage{
				Data: data,
			},
		}
	default:
		return fmt.Errorf("unknown system message type: %T", msg)
	}

	_, err := m.client.Send(ctx, req)
	if err != nil {
		return fmt.Errorf("system message via node dispatcher: %w", err)
	}
	return nil
}
