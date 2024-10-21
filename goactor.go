package goactor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hedisam/goactor/internal/intprocess"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/registry"
)

var (
	// ErrNamedActorNotFound is returned when an Actor cannot be found by a given name.
	ErrNamedActorNotFound = errors.New("no actor was found with the given name")
	// ErrNilDispatcher is returned with a nil dispatcher is used for sending a message
	ErrNilDispatcher = errors.New("cannot send message via a nil dispatcher")
	// ErrNilPID is returned when trying to send a message using a nil PID
	ErrNilPID = errors.New("cannot send message via a nil PID")
)

var (
	logger *slog.Logger
)

// ActorFactory is what expected by a supervisor WorkerSpec as well as a remote actor type registrar.
// It's necessary to have an ActorFactory that can create a fresh
// instance of the Actor whenever the supervisor restarts the process or when the node server spawns a new actor.
type ActorFactory func() Actor

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	var size int
	if sizeEnvVar := strings.TrimSpace(os.Getenv(registry.SizeEnvVar)); sizeEnvVar != "" {
		s, err := strconv.Atoi(sizeEnvVar)
		if err != nil {
			logger.Warn("Could not convert registry size env var value to int, using the default value",
				"error", err,
				slog.String("env_var", registry.SizeEnvVar),
				slog.Int("default_registry_size", registry.DefaultSize),
			)
			s = registry.DefaultSize
		}
		size = s
	}
	registry.InitRegistry(size)

	ns, err := startLocalNodeServer()
	if err != nil {
		logger.Error("Failed to enable clustering", "error", err)
		os.Exit(1)
	}
	initLocalNode(ns)
}

// SetLogHandler can be used to set a custom (non slog) log handler for the entire package.
// You should call this function in the beginning of your program. It is not safe to call it when you have
// active actors or supervisors. Access to the logger is not guarded by a mutex.
func SetLogHandler(h slog.Handler) {
	logger = slog.New(h)
}

// GetLogger can be used by internal packages to access the logger.
func GetLogger() *slog.Logger {
	return logger
}

// Dispatcher defines an interface used for sending messages.
type Dispatcher interface {
	// PID the process PID to send the message to.
	PID() *PID
}

// Named can be used to find and send a message to an actor registered by a name.
// It returns and invalid with PID if an actor with the given name can be found.
type Named string

// PID implements Dispatcher. It returns nil if no actor can be found with the given name.
func (named Named) PID() *PID {
	pid, found := registry.WhereIs(string(named))
	if !found {
		return nil
	}
	return &PID{
		internalPID: pid,
	}
}

// PID holds the internal pid. It is created when a process is created and can be used for interacting with it.
type PID struct {
	internalPID intprocess.PID
}

// PID implements Dispatcher.
func (pid *PID) PID() *PID {
	return pid
}

// Ref returns the process's unique reference ID.
func (pid *PID) Ref() string {
	return pid.internalPID.Ref()
}

// Spawn spawns the provided Actor and returns the corresponding Process Identifier.
// The provided actor can optionally implement ActorInitializer and ActorAfterFuncProvider interfaces.
func Spawn(ctx context.Context, actor Actor) (*PID, error) {
	var initFunc intprocess.InitFunc = func(ctx context.Context) error { return nil }
	var afterFunc intprocess.AfterFunc = func(ctx context.Context) error { return nil }
	var afterTimeout time.Duration

	if initializer, ok := actor.(ActorInitializer); ok {
		initFunc = initializer.Init
	}
	if afterFuncProvider, ok := actor.(ActorAfterFuncProvider); ok {
		var af AfterFunc
		afterTimeout, af = afterFuncProvider.AfterFunc()
		afterFunc = func(ctx context.Context) error {
			return af(ctx)
		}
	}

	pid, err := intprocess.SpawnLocal(
		ctx,
		logger,
		registry.GetRegistry(),
		initFunc,
		actor.Receive,
		afterFunc,
		afterTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("spawn local process: %w", err)
	}

	return &PID{
		internalPID: pid,
	}, nil
}

// Send sends a message to an ActorHandler with the provided PID.
func Send(ctx context.Context, d Dispatcher, msg any) error {
	if d == nil {
		return ErrNilDispatcher
	}
	pid := d.PID()
	if pid == nil {
		_, ok := d.(Named)
		if ok {
			return ErrNamedActorNotFound
		}
		return ErrNilPID
	}
	if pid.internalPID == nil {
		// should only happen if the PID has been retrieved using the Self function within a disposed actor.
		return registry.ErrSelfDisposed
	}

	err := mailbox.PushMessage(ctx, pid.internalPID.Dispatcher(), msg)
	if err != nil {
		return fmt.Errorf("push message via dispatcher: %w", err)
	}
	return nil
}
