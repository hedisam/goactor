package strategy

import "github.com/hedisam/goactor/supervisor/childstate"

type RestForOneStrategy struct {
	service supervisorService
}

func NewRestForOneStrategyHandler(service supervisorService) *RestForOneStrategy {
	return &RestForOneStrategy{
		service: service,
	}
}

func (s *RestForOneStrategy) Apply(_ *childstate.ChildState) error {
	// todo: implement rest_for_one strategy
	panic("rest_for_one -> to be implemented")
}
