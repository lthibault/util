package syncutil_test

import (
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
}
