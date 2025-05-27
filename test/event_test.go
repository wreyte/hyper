package event

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/wreyte/hyper/event"
)

func TestMultipleSubscribers(t *testing.T) {
	expect := 1
	topic := "foo.c"
	var (
		receivedValues []int
		mu             sync.Mutex
		wg             sync.WaitGroup
	)

	wg.Add(2)

	sub1 := event.Subscribe(topic, func(ctx context.Context, event any) {
		defer wg.Done()
		value, ok := event.(int)
		if !ok {
			t.Errorf("expected int got %v", reflect.TypeOf(event))
			return
		}
		mu.Lock()
		receivedValues = append(receivedValues, value)
		mu.Unlock()
	})

	sub2 := event.Subscribe(topic, func(ctx context.Context, event any) {
		defer wg.Done()
		value, ok := event.(int)
		if !ok {
			t.Errorf("expected int got %v", reflect.TypeOf(event))
			return
		}
		mu.Lock()
		receivedValues = append(receivedValues, value)
		mu.Unlock()
	})

	defer func() {
		event.Unsubscribe(sub1)
		event.Unsubscribe(sub2)
	}()

	event.Produce(topic, expect, event.HighPriority)

	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		t.Errorf("timeout waiting for event")
	case <-waitGroupDone(&wg):
		mu.Lock()
		defer mu.Unlock()
		if len(receivedValues) != 2 {
			t.Errorf("expected 2 received values, got %d", len(receivedValues))
		}
		for _, value := range receivedValues {
			if value != expect {
				t.Errorf("expected %d got %d", expect, value)
			}
		}
	}
}

func TestRaceCondition(t *testing.T) {
	expect := 1
	topic := "foo.d"
	var receivedValue int
	var wg sync.WaitGroup
	wg.Add(1)

	var sub event.Subscription
	done := make(chan struct{})

	go func() {
		sub = event.Subscribe(topic, func(ctx context.Context, event any) {
			defer wg.Done()
			value, ok := event.(int)
			if !ok {
				t.Errorf("expected int got %v", reflect.TypeOf(event))
			}
			receivedValue = value
		})
		close(done)
	}()

	<-done

	event.Produce(topic, expect, event.HighPriority)

	go func() {
		tmpSub := event.Subscribe(topic, func(ctx context.Context, event any) {})
		defer event.Unsubscribe(tmpSub)
	}()

	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		t.Errorf("timeout waiting for event")
	case <-waitGroupDone(&wg):
		if receivedValue != expect {
			t.Errorf("expected %d got %d", expect, receivedValue)
		}
	}

	event.Unsubscribe(sub)
}

func TestEventExecutionOrder(t *testing.T) {
	topic := "foo.order"
	var order []int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(4)

	sub := event.Subscribe(topic, func(ctx context.Context, event any) {
		defer wg.Done()
		value, ok := event.(int)
		if !ok {
			t.Errorf("expected int, got %T", event)
			return
		}
		mu.Lock()
		order = append(order, value)
		mu.Unlock()
	})

	defer event.Unsubscribe(sub)

	event.Produce(topic, 1, event.HighPriority)
	event.Produce(topic, 3, event.MediumPriority)
	event.Produce(topic, 2, event.LowPriority)
	event.Produce(topic, 4, event.HighPriority)

	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	select {
	case <-waitGroupDone(&wg):
		mu.Lock()
		defer mu.Unlock()
		t.Logf("execution order: %v", order)
		if len(order) != 4 {
			t.Errorf("expected 4 events, got %d", len(order))
		}
	case <-timeout.C:
		t.Errorf("timeout waiting for events")
	}
}

func TestUnsubscribeAfterEventProcessing(t *testing.T) {
	expect := 1
	topic := "foo.f"
	var receivedValue int
	var wg sync.WaitGroup
	wg.Add(1)

	sub := event.Subscribe(topic, func(ctx context.Context, event any) {
		defer wg.Done()
		value, ok := event.(int)
		if !ok {
			t.Errorf("expected int got %v", reflect.TypeOf(event))
		}
		receivedValue = value
	})

	event.Produce(topic, expect, event.HighPriority)

	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		t.Errorf("timeout waiting for event")
	case <-waitGroupDone(&wg):
		if receivedValue != expect {
			t.Errorf("expected %d got %d", expect, receivedValue)
		}
	}

	event.Unsubscribe(sub)
}

func waitGroupDone(wg *sync.WaitGroup) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}
