# hyper
- In-memory event bus that stores error logs
- Fire and forget implementation

## Installation
```
     go get github.com/wreyte/hyper
```

## Usage
import `github.com/wreyte/hyper/event`

- Check out the example.go file for a better understanding

```go

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

```
