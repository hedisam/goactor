package goactor

import (
	"fmt"

	"github.com/hedisam/goactor/internal/registry"
)

// Register associates a PID with the given name.
func Register(name string, pid *PID) error {
	// todo: add support for node actors
	return registry.RegisterNamed(name, pid.internalPID)
}

// Unregister disassociates a PID from the given name.
func Unregister(name string) {
	// todo: add support for node actors
	registry.UnregisterNamed(name)
}

// Link links self to the provided target PID.
// Link can only be called from the running Actor e.g. from the actor's Init, receive, or After functions.
func Link(linkee *PID) error {
	self, err := registry.Self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}

	err = self.Link(linkee.internalPID)
	if err != nil {
		return fmt.Errorf("link self to target linkee: %w", err)
	}
	return nil
}

// Unlink unlinks self from the linkee.
// Unlink can only be called from the running Actor e.g. from the actor's Init, receive, or After functions.
func Unlink(linkee *PID) error {
	self, err := registry.Self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}

	err = self.Unlink(linkee.internalPID)
	if err != nil {
		return fmt.Errorf("unlink self from linkee: %w", err)
	}
	return nil
}

// Monitor monitors the provided PID.
// The user defined receive function of monitor actors receive a sysmsg.Down message when a monitored actor goes down.
// Monitor can only be called from the running Actor e.g. from the actor's Init, receive, or After functions.
func Monitor(monitoree *PID) error {
	self, err := registry.Self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}

	err = self.Monitor(monitoree.internalPID)
	if err != nil {
		return fmt.Errorf("monitor target monitoree: %w", err)
	}
	return nil
}

// Demonitor de-monitors the provided PID.
// Demonitor can only be called from the running Actor e.g. from the actor's Init, receive, or After functions.
func Demonitor(monitoree *PID) error {
	self, err := registry.Self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}

	err = self.Demonitor(monitoree.internalPID)
	if err != nil {
		return fmt.Errorf("demonitor target monitoree: %w", err)
	}
	return nil
}

// SetTrapExit can be used to trap signals and exit messages from linked actors.
// A direct sysmsg.Signal with a sysmsg.ReasonKill cannot be trapped.
// SetTrapExit can only be called from the running Actor e.g. from the actor's Init, receive, or After functions.
func SetTrapExit(trapExit bool) error {
	self, err := registry.Self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}
	self.SetTrapExit(trapExit)
	return nil
}

// Self can be used from an actor process to retrieve the self *PID.
func Self() *PID {
	self, _ := registry.Self()
	return &PID{internalPID: self}
}
