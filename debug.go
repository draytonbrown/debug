package debug

import (
	"context"
	"errors"
	"fmt"

	de "github.com/draytonbrown/debug/errors"
)

// Command is a structure that represents a debug command
type Command[T any] struct {
	// Id is the debug code identifier.
	// eg. <service-name>.<endpoint-name>.<port>
	Id string `json:"Id"`
	// Payload is the debug command payload
	Payload T `json:"payload"`
}

type key struct{}

// enabled determines if the debug functionality is available.
// enabled is only enabled when `enable.go` is successfully added to the build.
var enabled bool

// NewCommand creates a new debug Command with the provided id and payload
func NewCommand[T any](id string, payload T) (*Command[T], error) {
	return &Command[T]{
		Id:      id,
		Payload: payload,
	}, nil
}

// Commands returns all the debug commands found in the provided Context
func Commands(ctx context.Context) ([]*Command[any], error) {
	if v := ctx.Value(key{}); v != nil {
		if c, ok := v.([]*Command[any]); ok {
			return c, nil
		}
	}
	return nil, de.NoCommandsFound
}

func command(ctx context.Context, id string) (*Command[any], error) {
	commands, err := Commands(ctx)
	if err != nil {
		return nil, de.CommandNotFound
	}

	for _, c := range commands {
		if c.Id == id {
			return c, nil
		}
	}

	return nil, de.CommandNotFound
}

// Context returns the provided context with a new debug command.
func Context(
	ctx context.Context,
	id string,
	payload any,
) (context.Context, error) {
	c, err := NewCommand(id, payload)
	if err != nil {
		return nil, err
	}

	commands, err := Commands(ctx)
	if err != nil {
		if !errors.Is(err, de.NoCommandsFound) {
			return nil, err
		}
	}
	commands = append(commands, c)

	return context.WithValue(ctx, key{}, commands), nil
}

// Wrap returns the result from the function `f` OR returns the debug payload.
// The function `f` is only executed if:
// - debug is not enabled
// - no debug command is found for the provided `id`
func Wrap[T any](
	ctx context.Context,
	id string,
	f func() (T, error),
) (T, error) {
	if !enabled {
		return f()
	}

	c, err := command(ctx, id)
	if err != nil {
		return f()
	}

	return func(cmd *Command[any]) (T, error) {
		switch p := (cmd.Payload).(type) {
		case T:
			return p, nil
		default:
			return *new(T), fmt.Errorf(
				"unexpected payload type: %T (wanted: %T)",
				cmd.Payload, *new(T),
			)
		}
	}(c)
}

// Do executes the function `f` if debug is enabled AND a debug Command is
// found with the provided `id`
func Do(ctx context.Context, id string, f func(*Command[any]) error) error {
	if !enabled {
		return nil
	}

	c, err := command(ctx, id)
	if err != nil {
		if errors.Is(err, de.CommandNotFound) {
			return nil
		}
		return err
	}

	return f(c)
}

// Custom returns the result from the function `f` OR returns the result
// of executing function `f2`. The function `f` is only executed if:
// - debug is not enabled
// - no debug command is found for the provided `id`
func Custom(
	ctx context.Context,
	id string,
	f func() error,
	f2 func(*Command[any]) error,
) error {
	if !enabled {
		return f()
	}

	c, err := command(ctx, id)
	if err != nil {
		return f()
	}
	return f2(c)
}
