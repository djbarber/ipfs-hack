package goprocessctx

import (
	"errors"
	"time"

	goprocess "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/goprocess"
	"github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"
)

const (
	closing = iota
	closed
)

type procContext struct {
	done  <-chan struct{}
	which int
}

// OnClosingContext derives a context from a given goprocess that will
// be 'Done' when the process is closing
func OnClosingContext(p goprocess.Process) context.Context {
	return &procContext{
		done:  p.Closing(),
		which: closing,
	}
}

// OnClosedContext derives a context from a given goprocess that will
// be 'Done' when the process is closed
func OnClosedContext(p goprocess.Process) context.Context {
	return &procContext{
		done:  p.Closed(),
		which: closed,
	}
}

func (c *procContext) Done() <-chan struct{} {
	return c.done
}

func (c *procContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c *procContext) Err() error {
	if c.which == closing {
		return errors.New("process closing")
	} else if c.which == closed {
		return errors.New("process closed")
	} else {
		panic("unrecognized process context type")
	}
}

func (c *procContext) Value(key interface{}) interface{} {
	return nil
}
