package strategy

import (
	"fmt"
	"iter"
	"slices"
	"time"
)

const (
	defaultMaxRestarts uint = 3
	defaultPeriod           = time.Second * 5
)

type strategyType string

const (
	// oneForOne If one child process terminates and is to be restarted, only that child process is affected.
	// This is the default restart strategy.
	oneForOne strategyType = ":one_for_one"

	// oneForAll If one child process terminates and is to be restarted,
	// all other child processes are terminated and then all child processes are restarted.
	oneForAll strategyType = ":one_for_all"

	// restForOne If one child process terminates and is to be restarted, the 'rest' of the child
	// processes (that is, the child processes after the terminated child process in the start order) are
	// terminated. Then the terminated child process and all child processes after it are restarted.
	restForOne strategyType = ":rest_for_one"
)

// ChildInfo holds information about the supervisor's children which is used by the strategy evaluator.
type ChildInfo struct {
	Name      string
	Temporary bool
	Stopped   bool
}

// Option defines an option function for configuring the strategy's restart intensity.
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

// Strategy holds strategy configuration for a supervisor.
// It determines how a supervisor should react to child termination events.
type Strategy struct {
	typ         strategyType
	period      time.Duration
	maxRestarts uint
}

// NewOneForOne returns a :one_for_one strategy configured with the provided restart intensity options.
func NewOneForOne(opts ...Option) *Strategy {
	return newStrategy(oneForOne, opts...)
}

// NewOneForAll returns a :one_for_all strategy configured with the provided restart intensity options.
func NewOneForAll(opts ...Option) *Strategy {
	return newStrategy(oneForAll, opts...)
}

// NewRestForOne returns a :rest_for_one strategy configured with the provided restart intensity options.
func NewRestForOne(opts ...Option) *Strategy {
	return newStrategy(restForOne, opts...)
}

// Default returns new instance of Strategy with the default options.
func Default() *Strategy {
	return newStrategy(oneForOne)
}

func newStrategy(typ strategyType, opts ...Option) *Strategy {
	s := &Strategy{
		typ:         typ,
		period:      defaultPeriod,
		maxRestarts: defaultMaxRestarts,
	}
	for opt := range slices.Values(opts) {
		opt(s)
	}
	return s
}

// MaxRestarts returns the max allowed restarts within the specified period.
func (s *Strategy) MaxRestarts() uint {
	return s.maxRestarts
}

// Period returns the period duration in which the restart intensity is evaluated and applied.
func (s *Strategy) Period() time.Duration {
	return s.period
}

// Evaluate evaluates the supervision strategy and determines which children to shut down and which to restart.
func (s *Strategy) Evaluate(terminated string, children []ChildInfo) (toShutdown []string, toRestart []string) {
	switch s.typ {
	case oneForOne:
		return nil, []string{terminated}
	case oneForAll:
		return evalOneForAll(children)
	case restForOne:
		return evalRestForOne(terminated, children)
	default:
		panic(fmt.Sprintf("unknown strategy type when evaluating: %q", s.typ))
	}
}

func evalOneForAll(children []ChildInfo) (toShutdown []string, toRestart []string) {
	for child := range reverse(children) {
		if !child.Stopped {
			toShutdown = append(toShutdown, child.Name)
		}
	}
	for child := range slices.Values(children) {
		if !child.Temporary {
			// a temporary child doesn't get to be restarted
			toRestart = append(toRestart, child.Name)
		}
	}
	return toShutdown, toRestart
}

func evalRestForOne(terminated string, children []ChildInfo) (toShutdown []string, toRestart []string) {
	for child := range reverse(children) {
		if child.Name == terminated {
			break
		}
		if !child.Stopped {
			toShutdown = append(toShutdown, child.Name)
		}
	}
	terminatedIdx := slices.IndexFunc(children, func(child ChildInfo) bool {
		return child.Name == terminated
	})
	for child := range slices.Values(children[terminatedIdx:]) {
		if !child.Temporary {
			toRestart = append(toRestart, child.Name)
		}
	}
	return toShutdown, toRestart
}

func reverse[T any](s []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := len(s) - 1; i >= 0; i-- {
			if !yield(s[i]) {
				return
			}
		}
	}
}
