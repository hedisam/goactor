package supervision

import (
	"fmt"
	"slices"
	"time"
)

// StrategyType determines how the supervisor restarts child actors.
type StrategyType uint

const (
	// StrategyOneForOne if a child process terminates, only that process is restarted
	StrategyOneForOne StrategyType = iota

	// StrategyOneForAll if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	StrategyOneForAll

	// StrategyRestForOne if a child process terminates, the terminated child process and
	// the rest of the specs started after it, are terminated and restarted.
	StrategyRestForOne
)

// Default values for supervision restart strategy.
const (
	defaultMaxRestarts uint = 3
	defaultPeriod           = time.Second * 5
)

// StrategyOption defines an option function for configuring Strategy.
type StrategyOption func(s *Strategy)

// StrategyWithMaxRestarts sets the strategy's max allowed restarts within the specified period.
func StrategyWithMaxRestarts(maxRestarts uint) StrategyOption {
	return func(s *Strategy) {
		s.maxRestarts = maxRestarts
	}
}

// StrategyWithPeriod sets the strategy's restart period.
func StrategyWithPeriod(duration time.Duration) StrategyOption {
	return func(s *Strategy) {
		if duration > 0 {
			s.period = duration
		}
	}
}

// Strategy holds the supervision strategy configuration which determines how the supervisor restarts child actors
// in the event of a child actor termination.
type Strategy struct {
	typ         StrategyType
	maxRestarts uint
	period      time.Duration
}

// NewStrategy returns a new supervision strategy.
func NewStrategy(strategyType StrategyType, opts ...StrategyOption) *Strategy {
	s := &Strategy{
		typ:         strategyType,
		maxRestarts: defaultMaxRestarts,
		period:      defaultPeriod,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func validateSupervisionStrategy(s *Strategy) error {
	validStrategyTypes := []StrategyType{StrategyOneForOne, StrategyOneForAll, StrategyRestForOne}
	if !slices.Contains(validStrategyTypes, s.typ) {
		return fmt.Errorf("invalid restart strategy type %q, valid strategies are: %q", s.typ, validStrategyTypes)
	}
	if s.period < 1 {
		return fmt.Errorf("invalid restarts period %d, period must be greater than 0", s.period)
	}
	return nil
}
