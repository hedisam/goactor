package supervisor

import (
	"fmt"
)

const (
	// if a child process terminates, only that process is restarted
	StrategyOptionOneForOne StrategyType = iota

	// if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	StrategyOptionOneForAll

	// if a child process terminates, the terminated child process and
	// the rest of the specs started after it, are terminated and restarted.
	StrategyOptionRestForOne
)

const (
	DefaultMaxRestarts int = 3
	DefaultPeriod      int = 5
)

type StrategyType int32

type Options struct {
	Strategy    StrategyType
	MaxRestarts int
	Period      int
}

func OneForOneStrategyOption() Options {
	return NewOptions(StrategyOptionOneForOne, DefaultMaxRestarts, DefaultPeriod)
}

func OneForAllStrategyOption() Options {
	return NewOptions(StrategyOptionOneForAll, DefaultMaxRestarts, DefaultPeriod)
}

func RestForOneStrategyOption() Options {
	return NewOptions(StrategyOptionRestForOne, DefaultMaxRestarts, DefaultPeriod)
}

func NewOptions(strategy StrategyType, maxRestarts, period int) Options {
	return Options{
		Strategy:    strategy,
		MaxRestarts: maxRestarts,
		Period:      period,
	}
}

func (opt *Options) validate() error {
	if opt.Strategy < 0 || opt.Strategy > 2 {
		return fmt.Errorf("invalid supervisor strategy: %d", opt.Strategy)
	} else if opt.Period < 1 {
		return fmt.Errorf("invalid restarts period - period must be greater than 0 - period: %d", opt.Period)
	} else if opt.MaxRestarts < 0 {
		return fmt.Errorf("invalid max restarts: %d", opt.MaxRestarts)
	}

	return nil
}
