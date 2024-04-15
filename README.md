# debug

![testing](https://github.com/draytonbrown/debug/actions/workflows/test.yaml/badge.svg)
[![godoc](https://godoc.org/github.com/draytonbrown/debug?status.svg)](https://godoc.org/github.com/draytonbrown/debug)

A small library to assist with testing.

## How To

Enable the debug library by setting the `ENVRIONMENT` variable to something other than `production` (case insensitive):

```shell
  > export ENVRIONMENT=development
```

Add a debug block to your production implementation:

```go
package main

import (
	"context"

	"github.com/draytonbrown/debug"

)

func Code(ctx context.Context) string {
	//...
	var code string
	_ := debug.Custom(ctx, "debug-id", func() error {
		// Production implementation...
		code = "production-code"
		return nil
	}, func(c *debug.Command[any]) error {
		// Testing implementation...
		code = c.Payload.(string)
		return nil
	})
	//...
	return code
}
```

Use a debug code in your tests to switch between production and debug implementations:

```go
package main

import (
	"context"
	"testing"

	"github.com/draytonbrown/debug"
)

func TestCode(t *testing.T) {
	ctx := context.Background()
	if c := Code(ctx); c != "production-code" {
		t.Errorf("wanted: production-code, got: %v", c)
	}

	ctx, _ = debug.Context(ctx, "debug-id", "debug-code")
	if c := Code(ctx); c != "debug-code" {
		t.Errorf("wanted: debug-code, got: %v", c)
	}
}
```

See the tests for more examples

## Thanks

This project was inspired by [Laurence Withers](https://github.com/lwithers)
