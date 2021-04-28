package goactor

import "fmt"

var ErrSendNilPID = fmt.Errorf("send failed: can not send message to a nil pid")
var ErrSendToSupervisor = fmt.Errorf("send failed: can not send message to a supervisor, use supervisor's supref instead")
var ErrSendNameNotFound = fmt.Errorf("send failed: no actor's been registered with the provided name")

var ErrLinkNilTargetPID = fmt.Errorf("failed to link: target pid is nil")
var ErrUnlinkNilTargetPID = fmt.Errorf("failed to unlink: target pid is nil")
var ErrMonitorNilTargetPID = fmt.Errorf("failed to monitor: target pid is nil")
var ErrDemonitorNilTargetPID = fmt.Errorf("failed to demonitor: target pid is nil")