package supervision

import (
	"fmt"
	"slices"
)

// StrategyType determines how the supervisor restarts child actors.
type StrategyType string

const (
	// StrategyOneForOne if a child process terminates, only that process is restarted
	StrategyOneForOne StrategyType = ":one_for_one"

	// StrategyOptionOneForAll if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	StrategyOptionOneForAll StrategyType = ":one_for_all"

	// StrategyOptionRestForOne if a child process terminates, the terminated child process and
	// the rest of the specs started after it, are terminated and restarted.
	StrategyOptionRestForOne StrategyType = ":rest_for_one"
)

// Default values for restart strategy.
const (
	DefaultMaxRestarts uint = 3
	DefaultPeriod      uint = 5
)

type StrategyOption func(s *Strategy)

func StrategyWithMaxRestarts(maxRestarts uint) StrategyOption {
	return func(s *Strategy) {
		s.MaxRestarts = maxRestarts
	}
}

func StrategyWithPeriod(period uint) StrategyOption {
	return func(s *Strategy) {
		s.Period = period
	}
}

// Strategy holds the supervision strategy configuration which determines how the supervisor restarts child actors
// in the event of a child actor termination.
type Strategy struct {
	Type        StrategyType
	MaxRestarts uint
	Period      uint
}

// NewStrategy creates a new supervision strategy.
func NewStrategy(strategyType StrategyType, opts ...StrategyOption) *Strategy {
	s := &Strategy{
		Type:        strategyType,
		MaxRestarts: DefaultMaxRestarts,
		Period:      DefaultPeriod,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Validate validates this RestartStrategy.
func (s *Strategy) Validate() error {
	validStrategyTypes := []StrategyType{StrategyOneForOne, StrategyOptionOneForAll, StrategyOptionRestForOne}
	if !slices.Contains(validStrategyTypes, s.Type) {
		return fmt.Errorf("invalid restart strategy type %q, valid strategies are: %q", s.Type, validStrategyTypes)
	}
	if s.Period < 1 {
		return fmt.Errorf("invalid restarts period %d, period must be greater than 0", s.Period)
	}
	return nil
}
