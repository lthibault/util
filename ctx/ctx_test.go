package ctxutil_test

import (
	"context"
	"testing"

	ctxutil "github.com/lthibault/util/ctx"
	"github.com/stretchr/testify/assert"
)

func TestC(t *testing.T) {
	t.Parallel()

	ch := make(chan struct{})
	close(ch)

	assert.ErrorIs(t, ctxutil.C(ch).Err(), context.Canceled,
		"error should be 'context.Canceled'")

	dl, ok := ctxutil.C(ch).Deadline()
	assert.False(t, ok,
		"deadline should not be set")
	assert.Zero(t, dl,
		"deadline should be zero-value 'time.Time'")
}
