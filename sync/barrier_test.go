package syncutil_test

import (
	"testing"
	"time"

	syncutil "github.com/lthibault/util/sync"
	"github.com/stretchr/testify/require"
)

func TestBarrierChan(t *testing.T) {
	const n = 10

	var (
		ran, finalized syncutil.Ctr
		bc             = syncutil.NewBarrierChan(n)
	)

	for i := 0; i < n; i++ {
		go func() {
			ran.Incr()
			bc.SignalAndWait(func() { finalized.Incr() })
		}()
	}

	select {
	case <-time.After(time.Millisecond * 10):
		t.Error("barrier never released")
	case <-bc.Done():
	}

	require.Equal(t, n, ran.Num())
	require.Equal(t, 1, finalized.Num())
}
