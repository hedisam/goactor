package strategy

import "github.com/hedisam/goactor/supervisor/childstate"

type OneForAllStrategy struct {
	service supervisorService
}

func NewOneForAllStrategyHandler(service supervisorService) *OneForAllStrategy {
	return &OneForAllStrategy{
		service: service,
	}
}

func (s *OneForAllStrategy) Apply(_ *childstate.ChildState) error {
	iterator := s.service.ChildrenIterator()
	for iterator.HasNext() {
		child := iterator.Value()
		err := child.Restart()
		if err != nil {
			return err
		}
	}
	return nil
}
