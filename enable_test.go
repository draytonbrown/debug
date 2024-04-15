package debug

import (
	"testing"
)

func TestMain(m *testing.M) {
	enabled = true
	m.Run()
}
