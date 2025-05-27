package event

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type eventStream struct {
	mu       sync.RWMutex
	subs     map[string][]Subscription
	highch   chan event
	mediumch chan event
	lowch    chan event
	quitch   chan struct{}
	workerWg sync.WaitGroup
}

func newStream() *eventStream {
	e := &eventStream{
		subs:     make(map[string][]Subscription),
		highch:   make(chan event, queueBufferSize),
		mediumch: make(chan event, queueBufferSize),
		lowch:    make(chan event, queueBufferSize),
		quitch:   make(chan struct{}),
	}
	go e.start()
	return e
}

func (e *eventStream) start() {
	ctx := context.Background()

	for i := 0; i < workersPerPriority; i++ {
		e.workerWg.Add(1)
		go e.worker(ctx, e.highch, HighPriority)
	}

	time.Sleep(10 * time.Millisecond)
	for i := 0; i < workersPerPriority; i++ {
		e.workerWg.Add(1)
		go e.worker(ctx, e.mediumch, MediumPriority)
	}

	time.Sleep(10 * time.Millisecond)
	for i := 0; i < workersPerPriority; i++ {
		e.workerWg.Add(1)
		go e.worker(ctx, e.lowch, LowPriority)
	}

	<-e.quitch
	slog.Info("Event stream received stop signal. Closing channels...")

	close(e.highch)
	close(e.mediumch)
	close(e.lowch)

	slog.Info("Waiting for all workers to finish...")
	e.workerWg.Wait()
	slog.Info("All event stream workers have shut down.")
}

func (e *eventStream) stop() {
	select {
	case <-e.quitch:
		slog.Warn("Stop already called on event stream.")
	default:
		close(e.quitch)
	}
}
