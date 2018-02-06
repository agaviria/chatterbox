package trace

import (
	"fmt"
	"io"
)

// Tracer interface describes an object capable of tracing events throughout code.
type Tracer interface {
	// Trace method will accept zero or more arguments of any type.
	Trace(...interface{})
}

// New creates a new Tracer that will write the output to the specified io.Writer.
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

// tracer is a Tracer that writes to an io.Writer.
type tracer struct {
	out io.Writer
}

// Trace writes the arguments to this Tracers io.Writer.
func (t *tracer) Trace(a ...interface{}) {
	fmt.Fprint(t.out, a...)
	fmt.Fprintln(t.out)
}
