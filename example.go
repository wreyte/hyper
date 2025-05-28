package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wreyte/hyper/event"
)

func main() {
	fmt.Println("\n Pub/Sub Usage Example")
	createUser := event.Subscribe("user_created", func(_ context.Context, msg any) {
		slog.Info(fmt.Sprintf("Received user_created event: %v", msg))
	})

	event.Produce("user_created", "Alice", event.HighPriority)
	event.Produce("user_created", "Bob", event.MediumPriority)
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n Unsubscribe Usage Example")
	event.Unsubscribe(createUser)
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n Non deterministic but should mostly return HighPriority first")
	mixedPriority := event.Subscribe("mixed_priority", func(_ context.Context, msg any) {
		slog.Info(fmt.Sprintf("Mixed Priority Handler received: %v", msg))
	})

	event.Produce("mixed_priority", "Low Priority Message", event.LowPriority)
	event.Produce("mixed_priority", "High Priority Message", event.HighPriority)
	event.Produce("mixed_priority", "Medium Priority Message", event.MediumPriority)
	event.Produce("mixed_priority", "Another High Priority Message", event.HighPriority)
	time.Sleep(100 * time.Millisecond)

	event.Unsubscribe(mixedPriority)

	fmt.Println("\n No Subscriber for this handler")
	fmt.Println("Should receive an error log in the json file")
	event.Produce("non_existent_topic", "Some message", event.MediumPriority)
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n Intentionally adding a panic to handler")
	fmt.Println("Should also receive an error log in the json file")
	_ = event.Subscribe("panic_topic", func(_ context.Context, msg any) {
		slog.Info(fmt.Sprintf("Handler for panic_topic received: %v", msg))
		panic("Oops! Something went wrong in the handler!")
	})
	event.Produce("panic_topic", "Faulty data", event.HighPriority)
	time.Sleep(100 * time.Millisecond)

	event.Stop()
}
