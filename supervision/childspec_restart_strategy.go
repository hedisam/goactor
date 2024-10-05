package supervision

import (
	"fmt"
	"slices"
)

const (
	// RestartAlways means the child actor is always restarted (the default).
	RestartAlways RestartStrategy = ":permanent"
	// RestartTransient means the child actor is restarted only if it terminates abnormally (with an exit reason
	// other than :normal and :shutdown)
	RestartTransient RestartStrategy = ":transient"
	// RestartNever means the child actor is never restarted, regardless of its termination reason.
	RestartNever RestartStrategy = ":temporary"
)

// RestartStrategy determines when to restart a child actor if it terminates.
type RestartStrategy string

func validateRestartStrategy(s RestartStrategy) error {
	validStrategies := []string{
		string(RestartAlways),
		string(RestartTransient),
		string(RestartNever),
	}
	if !slices.Contains(validStrategies, string(s)) {
		return fmt.Errorf("invalid child restart strategy %q, valid restart strategies are: [%s]", s, validStrategies)
	}
	return nil
}
