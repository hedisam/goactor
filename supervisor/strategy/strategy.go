package strategy

import "github.com/hedisam/goactor/supervisor/childstate"

type supervisorService interface {
	ChildrenIterator() *childstate.ChildrenStateIterator
}
