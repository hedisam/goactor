package goactor

import (
	"context"
	"errors"
	"fmt"

	"github.com/hedisam/goactor/sysmsg"
)

type systemPID struct {
	pid *PID
}

// PushSystemMessage sends a system message to the underlying PID.
func (p *systemPID) PushSystemMessage(ctx context.Context, msg *sysmsg.Message) error {
	return sendSystemMessage(ctx, p.pid, msg)
}

// ID returns the ID of the underlying PID.
func (p *systemPID) ID() string {
	return p.pid.ID()
}

func sendSystemMessage(ctx context.Context, pid ProcessIdentifier, msg *sysmsg.Message) error {
	if pid.PID() == nil {
		return errors.New("cannot send message via a nil PID")
	}

	err := pid.PID().dispatcher.PushSystemMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("push system message via dispatcher: %w", err)
	}
	return nil
}
