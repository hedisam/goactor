package strategy

import "github.com/hedisam/goactor/supervisor/childstate"

type OneForOneStrategy struct{}

func NewOneForOneStrategyHandler() *OneForOneStrategy {
	return &OneForOneStrategy{}
}

func (s *OneForOneStrategy) Apply(child *childstate.ChildState) error {
	return child.Restart()
}
