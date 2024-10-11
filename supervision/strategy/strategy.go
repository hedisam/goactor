package strategy

import (
	"fmt"
	"slices"
	"time"
)

const (
	defaultMaxRestarts uint = 3
	defaultPeriod           = time.Second * 5
)

// Type represents the strategy type which determines how the supervisor restarts child actors
// in relation to the other children.
type Type uint

const (
	// OneForOne if a child process terminates, only that process is restarted
	OneForOne Type = iota

	// OneForAll if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	OneForAll

	// RestForOne if a child process terminates, the terminated child process and
	// the rest of the specs started after it, are terminated and restarted.
	RestForOne
)

// String returns the string representation of a strategy Type.
func (t Type) String() string {
	switch t {
	case OneForOne:
		return "ONE_FOR_ONE"
	case OneForAll:
		return "ONE_FOR_ALL"
	case RestForOne:
		return "REST_FOR_ONE"
	default:
		return fmt.Sprintf("UNKNOWN_STRATEGY_TYPE:%d", t)
	}
}

// Strategy holds the configuration for a supervisor strategy.
type Strategy struct {
	typ         Type
	maxRestarts uint
	period      time.Duration
}

// Type returns the configured type of this strategy.
func (s *Strategy) Type() Type {
	return s.typ
}

// MaxRestarts returns the configured max restarts value of this strategy.
func (s *Strategy) MaxRestarts() uint {
	return s.maxRestarts
}

// Period returns the configured period duration of this strategy.
func (s *Strategy) Period() time.Duration {
	return s.period
}

// NewOneForOne returns a new instance of Strategy with type set to OneForOne.
func NewOneForOne(opts ...Option) *Strategy {
	s := &Strategy{
		typ:         OneForOne,
		maxRestarts: defaultMaxRestarts,
		period:      defaultPeriod,
	}
	for opt := range slices.Values(opts) {
		opt(s)
	}
	return s
}

// NewOneForAll returns a new instance of Strategy with type set to OneForAll.
func NewOneForAll(opts ...Option) *Strategy {
	s := &Strategy{
		typ:         OneForAll,
		maxRestarts: defaultMaxRestarts,
		period:      defaultPeriod,
	}
	for opt := range slices.Values(opts) {
		opt(s)
	}
	return s
}

// NewRestForOne returns a new instance of Strategy with type set to RestForOne.
func NewRestForOne(opts ...Option) *Strategy {
	s := &Strategy{
		typ:         RestForOne,
		maxRestarts: defaultMaxRestarts,
		period:      defaultPeriod,
	}
	for opt := range slices.Values(opts) {
		opt(s)
	}
	return s
}

// Option defines an option function for configuring an instance of strategy.
type Option func(s *Strategy)

// WithMaxRestarts sets the strategy's max allowed restarts within the specified period.
func WithMaxRestarts(maxRestarts uint) Option {
	return func(s *Strategy) {
		s.maxRestarts = maxRestarts
	}
}

// WithPeriod sets the strategy's restart period.
func WithPeriod(duration time.Duration) Option {
	return func(s *Strategy) {
		if duration > 0 {
			s.period = duration
		}
	}
}
