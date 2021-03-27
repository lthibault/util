package syncutil_test

import (
	"context"
	"errors"
	"testing"
	"time"

	syncutil "github.com/lthibault/util/sync"
	"github.com/stretchr/testify/assert"
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

	var g syncutil.FuncGroup

	var ctr syncutil.Ctr
	for i := 0; i < 10; i++ {
		g.Go(func() {
			time.Sleep(time.Millisecond * time.Duration(i))
			ctr.Incr()
		})
	}

	g.Wait()
	assert.Equal(t, 10, ctr.Num())
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
	assert.Equal(t, 1, ctr.Num())

	ctr.Decr()
	assert.Zero(t, ctr)
}
