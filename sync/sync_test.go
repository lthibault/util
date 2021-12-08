package syncutil_test

import (
	"context"
	"errors"
	"testing"
	"time"

	syncutil "github.com/lthibault/util/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAny(t *testing.T) {
	t.Parallel()

	t.Run("Nop", func(t *testing.T) {
		t.Parallel()

		var any syncutil.Any
		assert.NoError(t, any.Wait())
	})

	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		var any syncutil.Any
		any.Go(func() error { return errors.New("test") })
		any.Go(func() error { return nil })
		assert.NoError(t, any.Wait())
	})

	t.Run("Fail", func(t *testing.T) {
		t.Parallel()

		var any syncutil.Any
		any.Go(func() error { return errors.New("test") })
		assert.Error(t, any.Wait())
	})

	t.Run("Context", func(t *testing.T) {
		t.Parallel()

		var flag syncutil.Flag
		any, ctx := syncutil.AnyWithContext(context.Background())

		any.Go(func() error {
			time.Sleep(time.Millisecond * 5)
			return nil
		})

		any.Go(func() error {
			select {
			case <-time.After(time.Second):
				t.Error("unreachable")
			case <-ctx.Done():
				flag.Set()
			}

			return errors.New("test")
		})

		assert.NoError(t, any.Wait())
		assert.Eventually(t, flag.Bool, time.Millisecond, time.Microsecond*100)
	})
}

func TestFuncGroup(t *testing.T) {
	t.Parallel()

	const n = 10

	var (
		g   syncutil.FuncGroup
		ctr syncutil.Ctr
	)

	for i := 0; i < n; i++ {
		g.Go(func(i int) func() {
			return func() {
				time.Sleep(time.Millisecond * time.Duration(i))
				ctr.Incr()
			}
		}(i))
	}

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		g.Wait()
	}()

	select {
	case <-ch:
		assert.Equal(t, n, ctr.Int())
	case <-time.After(time.Millisecond*n + (n / 2)):
		t.Error("regression:  is FuncGroup actually spawning a goroutine?")
	}
}

func TestFlag(t *testing.T) {
	t.Parallel()

	var flag syncutil.Flag

	assert.False(t, flag.Bool())

	flag.Set()
	assert.True(t, flag.Bool())

	flag.Unset()
	assert.False(t, flag.Bool())
}

func TestCtr(t *testing.T) {
	t.Parallel()

	var ctr syncutil.Ctr

	assert.Zero(t, ctr)

	ctr.Incr()
	assert.Equal(t, 1, ctr.Int())

	ctr.Decr()
	assert.Zero(t, ctr)
}

func TestJoin(t *testing.T) {
	t.Parallel()

	var (
		j   syncutil.Join
		ctr syncutil.Ctr
	)

	j.Go(func() error { return errors.New("test") })
	j.Go(func() error {
		time.Sleep(time.Millisecond)
		ctr.Incr()
		return nil
	})

	require.EqualError(t, j.Wait(), "test")
}
