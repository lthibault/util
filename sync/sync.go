// Package syncutil contains advanced synchronization primitives.
package syncutil

import (
	"context"
	"sync"
	"sync/atomic"

	"go.uber.org/multierr"
)

// FuncGroup calls a group of functions in separate goroutines and waits until they have
// all returned.  Compare to https://pkg.go.dev/golang.org/x/sync/errgroup.
//
// A zero-value FuncGroup is valid.  FuncGroup must not be copied after first use.
type FuncGroup sync.WaitGroup

// Go runs the supplied function in a goroutine
func (g *FuncGroup) Go(f func()) {
	(*sync.WaitGroup)(g).Add(1)
	go func() {
		defer (*sync.WaitGroup)(g).Done()
		f()
	}()
}

// Wait for all goroutines to complete
func (g *FuncGroup) Wait() { (*sync.WaitGroup)(g).Wait() }

// Any calls a group of functions in separate goroutines, and checks that at least one
// function call succeeded.  The Wait method returns an error if (and only if) all
// calls to Go have failed.  Compare to https://pkg.go.dev/golang.org/x/sync/errgroup.
//
// A zero-value Any is valid.  Any must not be copied after first use.
type Any struct {
	cancel context.CancelFunc

	wg sync.WaitGroup

	ok      Flag
	errOnce sync.Once
	err     error
}

// AnyWithContext returns a new Any and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// succeeds (returns a nil error) or the first time Wait returns, whichever occurs
// first.
func AnyWithContext(ctx context.Context) (*Any, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Any{cancel: cancel}, ctx
}

// Go calls the given function in a new goroutine.
//
// The first call to return a nil error cancels the group; wait will then return nil.
func (a *Any) Go(f func() error) {
	a.wg.Add(1)
	defer a.wg.Done()

	if err := f(); err != nil {
		a.errOnce.Do(func() { a.err = err })
		return
	}

	if a.cancel != nil {
		a.cancel()
	}

	a.ok.Set()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns nil if any of the calls succeeded.  Otherwise, it returns the first non-nil
// error encountered.
func (a *Any) Wait() error {
	a.wg.Wait()

	if a.cancel != nil {
		a.cancel()
	}

	if a.ok.Bool() {
		return nil
	}

	return a.err
}

// Join calls a group of functions in separate goroutines and waits until they have
// all returned.  The error returned by Join is a go.uber.org/multierr and is non-nil
// if any function returned a non-nil error.
//
// Unlike pkg.go.dev/golang.org/x/sync/errgroup, Join does not return until the full
// set uf goroutines has finished
//
// A zero-value Join is valid.  Join must not be copied after first use.
type Join struct {
	fg FuncGroup

	mu  sync.Mutex
	err error
}

func (j *Join) Go(f func() error) {
	j.fg.Go(func() {
		if err := f(); err != nil {
			j.mu.Lock()
			multierr.AppendInto(&j.err, err)
			j.mu.Unlock()
		}
	})
}

func (j *Join) Wait() error {
	j.fg.Wait()
	return j.err
}

// Ctr is a lock-free counter
type Ctr uint32

// Incr increments the counter
func (ctr *Ctr) Incr() uint32 { return atomic.AddUint32((*uint32)(ctr), 1) }

// Decr decrements the counter
func (ctr *Ctr) Decr() uint32 { return atomic.AddUint32((*uint32)(ctr), ^uint32(0)) }

// Num returns the generic-integer value of the counter.
// This is useful for integer comparisons, e.g. with `len`.
func (ctr *Ctr) Num() int { return int(atomic.LoadUint32((*uint32)(ctr))) }

// Flag is a lock-free boolean flag
type Flag uint32

// Set the flag's value to true
func (f *Flag) Set() { atomic.CompareAndSwapUint32((*uint32)(f), 0, 1) }

// Unset the flag, making its value false
func (f *Flag) Unset() { atomic.CompareAndSwapUint32((*uint32)(f), 1, 0) }

// Bool evaluates the flag's value
func (f *Flag) Bool() bool { return atomic.LoadUint32((*uint32)(f)) != 0 }
