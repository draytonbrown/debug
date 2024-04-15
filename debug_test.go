package debug_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/draytonbrown/debug"
)

func TestContextAndCommands(t *testing.T) {
	ctx := context.Background()

	id := "debug-id-1"
	payload := []string{"string1", "string2"}

	var err error
	ctx, err = debug.Context(ctx, id, payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	commands, err := debug.Commands(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var found bool
	for _, cmd := range commands {
		if cmd.Id != id {
			continue
		}
		found = true
		if !reflect.DeepEqual(cmd.Payload, payload) {
			t.Errorf("got %v, want %v", cmd, payload)
		}
	}
	if !found {
		t.Errorf("did not find %q in commands", id)
	}
}

func TestWrap(t *testing.T) {
	ctx := context.Background()

	t.Run("unknown debug code - OK", func(t *testing.T) {
		s := "Production implementation called"
		v, err := debug.Wrap(ctx, "unknown", func() (*string, error) {
			return &s, nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if strings.Compare(*v, s) != 0 {
			t.Errorf("got %v, want %v", *v, s)
		}
	})

	t.Run("unknown debug code - NOK", func(t *testing.T) {
		expErr := errors.New("production implementation error")
		_, err := debug.Wrap(ctx, "unknown", func() (*string, error) {
			return nil, expErr
		})
		if !errors.Is(err, expErr) {
			t.Errorf("unexpected error: %v", err)
		}
	})

	id := "debug-id-1"
	payload := []string{"debug-string1", "debug-string2"}

	var err error
	ctx, err = debug.Context(ctx, id, payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Run("OK", func(t *testing.T) {
		v, err := debug.Wrap[[]string](ctx, id, func() ([]string, error) {
			t.Errorf("unexpected non-debug code executed")
			v := []string{"string1", "string2"}
			return v, nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(v, payload) {
			t.Errorf("unexpected result, got %v, want %v", v, payload)
		}
	})

	t.Run("invalid debug payload type", func(t *testing.T) {
		ctx = context.Background()
		ctx, err = debug.Context(ctx, "invalid", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		_, err := debug.Wrap(ctx, "invalid", func() (string, error) {
			t.Errorf("unexpected non-debug code executed")
			return "", nil
		})
		if err == nil {
			t.Errorf("unexpected nil error")
		}
		if err.Error() != "unexpected payload type: <nil> (wanted: string)" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDo(t *testing.T) {
	ctx := context.Background()

	id := "debug-id-1"
	payload := []string{"debug-string1", "debug-string2"}

	var err error
	ctx, err = debug.Context(ctx, id, payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Run("unknown debug code", func(t *testing.T) {
		err := debug.Do(ctx, "unknown", func(c *debug.Command[any]) error {
			t.Errorf("unexpected debug code executed")
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("debug code found", func(t *testing.T) {
		var count int
		err := debug.Do(ctx, "debug-id-1", func(c *debug.Command[any]) error {
			count++
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if count != 1 {
			t.Errorf("do execution count, got %v, want %v", count, 1)
		}
	})

	t.Run("debug code found, return error", func(t *testing.T) {
		expErr := errors.New("production implementation error")
		err := debug.Do(ctx, "debug-id-1", func(c *debug.Command[any]) error {
			return expErr
		})
		if !errors.Is(err, expErr) {
			t.Errorf("unexpected error, got: %v, wanted: %v", err, expErr)
		}
	})
}

func TestCustom(t *testing.T) {
	ctx := context.Background()

	id := "debug-id-1"
	payload := []string{"debug-string1", "debug-string2"}

	var err error
	ctx, err = debug.Context(ctx, id, payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Run("unknown debug code, successful response", func(t *testing.T) {
		var exp string
		err := debug.Custom(ctx, "unknown", func() error {
			exp = "The production implementation"
			return nil
		}, func(c *debug.Command[any]) error {
			t.Errorf("unexpected debug code executed")
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if strings.Compare(exp, "The production implementation") != 0 {
			t.Errorf("got %v, want %v", "The production implementation", exp)
		}
	})

	t.Run("unknown debug code, error response", func(t *testing.T) {
		expErr := errors.New("production implementation error")
		err = debug.Custom(ctx, "unknown", func() error {
			return expErr
		}, func(c *debug.Command[any]) error {
			t.Errorf("unexpected debug code executed")
			return nil
		})
		if !errors.Is(err, expErr) {
			t.Errorf("unexpected error got %v, want %v", err, expErr)
		}
	})

	t.Run("debug code found, successful response", func(t *testing.T) {
		var exp string
		err := debug.Custom(ctx, id, func() error {
			t.Errorf("unexpected debug code executed")
			return nil
		}, func(c *debug.Command[any]) error {
			exp = "The debug implementation"
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if strings.Compare(exp, "The debug implementation") != 0 {
			t.Errorf("got %v, want %v", exp, "The debug implementation")
		}
	})

	t.Run("debug code found, error response", func(t *testing.T) {
		expErr := errors.New("production implementation error")
		err := debug.Custom(ctx, id, func() error {
			t.Errorf("unexpected debug code executed")
			return nil
		}, func(c *debug.Command[any]) error {
			return expErr
		})
		if !errors.Is(err, expErr) {
			t.Errorf("unexpected error got %v, want %v", err, expErr)
		}
	})

}
