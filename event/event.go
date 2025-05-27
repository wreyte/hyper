package event

import (
	"context"
	"log/slog"
	"os"
)

type HandlerFunc func(context.Context, any)

type Priority int

const (
	HighPriority Priority = iota
	MediumPriority
	LowPriority
)

const (
	queueBufferSize    = 512
	workersPerPriority = 8
	deadLetterLogFile  = "dead_letter_logs.json"
)

var (
	stream               *eventStream
	deadLetterLogger     *slog.Logger
	deadLetterFileHandle *os.File
)

func Produce(topic string, event any, priority Priority) {
	stream.emit(topic, event, priority)
}

func Subscribe(topic string, handler HandlerFunc) Subscription {
	return stream.subscribe(topic, handler)
}

func Unsubscribe(sub Subscription) {
	stream.unsubscribe(sub)
}

func Stop() {
	if stream != nil {
		stream.stop()
	}

	if deadLetterFileHandle != nil {
		err := deadLetterFileHandle.Close()
		if err != nil {
			slog.Error("Failed to close dead letter log file",
				"error", err)
		}
	}
}

func init() {
	var err error
	deadLetterFileHandle, err = os.OpenFile(deadLetterLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Error("Failed to open dead letter log file, falling back to stderr", "error", err)
		deadLetterLogger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	} else {
		deadLetterLogger = slog.New(slog.NewJSONHandler(deadLetterFileHandle, nil))
	}

	stream = newStream()
}
