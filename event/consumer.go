package event

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"
)

func (e *eventStream) worker(ctx context.Context, ch <-chan event, priority Priority) {
	defer e.workerWg.Done()
	slog.Info("Starting event worker", "priority", priority)

	for evt := range ch {
		e.processEvent(ctx, evt)
	}
	slog.Info("Event worker stopped", "priority", priority)
}

func (e *eventStream) processEvent(ctx context.Context, evt event) {
	e.mu.RLock()
	handlers, ok := e.subs[evt.topic]
	if !ok {
		e.mu.RUnlock()
		slog.Warn("No subscribers for topic",
			"topic", evt.topic,
			"event_message", evt.message)
		deadLetterLogger.Error("No subscribers for topic",
			"topic", evt.topic,
			"event_message", evt.message)
		return
	}

	copiedHandlers := make([]Subscription, len(handlers))
	copy(copiedHandlers, handlers)
	e.mu.RUnlock()

	for _, sub := range copiedHandlers {
		go func(s Subscription, eventMsg any) {
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("panic in event handler for topic %s: %v", s.Topic, r)
					slog.Error("Event handler panicked",
						"topic", s.Topic,
						"error", err,
						"event_message", eventMsg)
					deadLetterLogger.Error("Event handler panicked",
						"topic", s.Topic,
						"error", err,
						"event_message", eventMsg)
				}
			}()
			s.Fn(ctx, eventMsg)
		}(sub, evt.message)
	}
}

func (e *eventStream) subscribe(topic string, handler HandlerFunc) Subscription {
	e.mu.Lock()
	defer e.mu.Unlock()

	sub := Subscription{
		CreatedAt: time.Now().UnixNano(),
		Topic:     topic,
		Fn:        handler,
	}

	if _, ok := e.subs[topic]; !ok {
		e.subs[topic] = []Subscription{}
	}

	e.subs[topic] = append(e.subs[topic], sub)
	slog.Info("Subscribed handler to topic",
		"topic", topic)

	return sub
}

func (e *eventStream) unsubscribe(sub Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if handlers, ok := e.subs[sub.Topic]; ok {
		initialLen := len(handlers)
		e.subs[sub.Topic] = slices.DeleteFunc(handlers, func(s Subscription) bool {
			return sub.CreatedAt == s.CreatedAt
		})

		if len(e.subs[sub.Topic]) == 0 {
			delete(e.subs, sub.Topic)
			slog.Info("Removed topic as no more subscribers",
				"topic", sub.Topic)
		}

		if len(e.subs[sub.Topic]) < initialLen {
			slog.Info("Unsubscribed handler from topic",
				"topic", sub.Topic,
				"created_at", sub.CreatedAt)
		} else {
			slog.Warn("Attempted to unsubscribe a non-existent subscription",
				"topic", sub.Topic,
				"created_at", sub.CreatedAt)
		}
	} else {
		slog.Warn("Attempted to unsubscribe from a topic with no subscribers",
			"topic", sub.Topic)
	}
}
