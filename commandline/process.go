package commandline

import (
	"io"
	"os"
)

type Signaller interface {
	Signal(os.Signal) error
}

type Response struct {
	Signaller
	Stdout io.Writer
}

// A Handler called with a line read from the command line.
type Handler interface {
	Handle(Response, string)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as command line handlers. If f is a
// function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc func(Response, string)

// Handle calls f(w, r).
func (f HandlerFunc) Handle(r Response, l string) {
	f(r, l)
}
