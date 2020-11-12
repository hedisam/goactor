package sysmsg

import (
	"github.com/hedisam/goactor/internal/intlpid"
)

type SystemMessage interface {
	Sender() intlpid.InternalPID
	Reason() interface{}
	Origin() SystemMessage
}
