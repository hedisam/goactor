package goactor

import "fmt"

var ErrSendNilPID = fmt.Errorf("send failed: can not send message to a nil pid")
var ErrSendToSupervisor = fmt.Errorf("send failed: can not send message to a supervisor, use supervisor's supref instead")
var ErrSendNameNotFound = fmt.Errorf("send failed: no actor's been registered with the provided name")