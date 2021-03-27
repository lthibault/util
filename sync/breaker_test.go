package syncutil_test

import (
	"context"
	"errors"
	"testing"
	"time"

	syncutil "github.com/lthibault/util/sync"
	"github.com/stretchr/testify/assert"
)

func TestBreaker(t *testing.T) {
	t.Parallel()

	t.Run("Nop", func(t *testing.T) {
		t.Parallel()

		var b syncutil.Breaker

		b.Break()
		assert.NoError(t, b.Wait())
	})

	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		var b syncutil.Breaker

		b.Go(func() error { return nil })
		assert.NoError(t, b.Wait())
	})

	t.Run("Fail", func(t *testing.T) {
		t.Parallel()

		var b syncutil.Breaker

		b.Go(func() error { return errors.New("test") })
		b.Break()
		assert.Error(t, b.Wait())
	})

	t.Run("Break", func(t *testing.T) {
		t.Parallel()

		var b syncutil.Breaker

		b.Go(func() error {
			time.Sleep(time.Millisecond * 10)
			return errors.New("test")
		})

		b.Break()
		b.Go(func() error { panic("unreachable") }) // expect skip

		assert.Error(t, b.Wait())
	})

	t.Run("Context", func(t *testing.T) {
		t.Parallel()

		t.Run("Nop", func(t *testing.T) {
			t.Parallel()

			b, ctx := syncutil.BreakerWithContext(context.Background())
			b.Break()
			assert.ErrorIs(t, ctx.Err(), context.Canceled)
		})

		t.Run("Func", func(t *testing.T) {
			t.Parallel()

			var flag syncutil.Flag
			b, ctx := syncutil.BreakerWithContext(context.Background())

			b.Go(func() error {
				time.Sleep(time.Millisecond * 5)
				return nil
			})

			b.Go(func() error {
				select {
				case <-time.After(time.Second):
					t.Error("unreachable")
				case <-ctx.Done():
					flag.Set()
				}

				return errors.New("test")
			})

			assert.NoError(t, b.Wait())
			assert.Eventually(t, flag.Bool, time.Millisecond, time.Microsecond*100)
		})
	})
}
