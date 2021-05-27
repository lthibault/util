package syncutil

import "sync/atomic"

type Barrier uint32

func (b *Barrier) Add(i uint32) {
	atomic.AddUint32((*uint32)(b), i)
}

func (b *Barrier) Reset() {
	atomic.StoreUint32((*uint32)(b), 0)
}

func (b *Barrier) Ready(finalize func()) (ready bool) {
	if ready = atomic.AddUint32((*uint32)(b), ^uint32(0)) == 0; ready {
		finalize()
	}

	return
}

type BarrierChan struct {
	b  *Barrier
	cq chan struct{}
}

func NewBarrierChan(n uint32) BarrierChan {
	return BarrierChan{
		b:  (*Barrier)(&n),
		cq: make(chan struct{}),
	}
}

func (b BarrierChan) Done() <-chan struct{} { return b.cq }

func (b BarrierChan) SignalAndWait(finalize func()) {
	b.b.Ready(func() {
		defer close(b.cq)
		finalize()
	})
	<-b.cq
}
