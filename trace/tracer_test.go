package trace

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	// buf will be used to capture output of Tracer.
	var b bytes.Buffer
	tracer := New(&b)

	if tracer == nil {
		t.Error("Return from New should not be nil")
	} else {
		tracer.Trace("Hello from trace pkg.")
		if b.String() != "Hello from trace pkg.\n" {
			t.Errorf("Trace mismatch: %s", b.String())
		}
	}
}

// TestOff will test Mute() on Trace object.
func TestOff(t *testing.T) {
	var silentTracer Tracer = Mute()
	silentTracer.Trace("something")
}
