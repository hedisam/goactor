package supervision

import (
	"fmt"
	"slices"
)

const (
	// Permanent A permanent child process is always restarted
	Permanent RestartType = ":permanent"
	// Transient A transient child process is restarted only if it terminates abnormally, that is, with another exit
	// reason than sysmsg.ReasonNormal, sysmsg.ReasonShutdown.
	Transient RestartType = ":transient"
	// Temporary A temporary child process is never restarted (even when the supervisor's restart strategy is
	// :rest_for_one or :one_for_all and a sibling's death causes the temporary process to be terminated).
	Temporary RestartType = ":temporary"
)

// RestartType defines when a terminated child process must be restarted.
type RestartType string

func validateRestartType(s RestartType) error {
	validTypes := []string{
		string(Permanent),
		string(Transient),
		string(Temporary),
	}
	if !slices.Contains(validTypes, string(s)) {
		return fmt.Errorf("invalid child restart type %q, valid restart types are: [%s]", s, validTypes)
	}
	return nil
}
