package syncutil_test

import (
	"testing"
	"time"

	syncutil "github.com/lthibault/util/sync"
	"github.com/stretchr/testify/require"
)

func TestBarrierChan(t *testing.T) {
	var (
		ctr syncutil.Ctr
		bc  = syncutil.NewBarrierChan(2)
	)

	go bc.SignalAndWait(func() { ctr.Incr() })
	go bc.SignalAndWait(func() { ctr.Incr() })

	select {
	case <-time.After(time.Millisecond * 10):
		t.Error("barrier never released")
	case <-bc.Done():
	}

	require.Equal(t, 1, ctr.Num())
}
