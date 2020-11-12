package supervisor

import (
	"fmt"
)

const (
	// if a child process terminates, only that process is restarted
	OneForOneStrategy Strategy = iota

	// if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	OneForAllStrategy

	// if a child process terminates, the terminated child process and
	// the rest of the specs started after it, are terminated and restarted.
	RestForOneStrategy
)

const (
	DefaultMaxRestarts int = 3
	DefaultPeriod      int = 5
)

type Strategy int32

type Options struct {
	Strategy    Strategy
	MaxRestarts int
	Period      int
}

func OneForOneStrategyOption() Options {
	return NewOptions(OneForOneStrategy, DefaultMaxRestarts, DefaultPeriod)
}

func OneForAllStrategyOption() Options {
	return NewOptions(OneForAllStrategy, DefaultMaxRestarts, DefaultPeriod)
}

func RestForOneStrategyOption() Options {
	return NewOptions(RestForOneStrategy, DefaultMaxRestarts, DefaultPeriod)
}

func NewOptions(strategy Strategy, maxRestarts, period int) Options {
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
