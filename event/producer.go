package event

import (
	"log/slog"
)

func (e *eventStream) emit(topic string, v any, priority Priority) {
	evt := event{topic: topic, message: v, priority: priority}

	select {
	case <-e.quitch:
		slog.Warn("Attempted to emit event after stream stop signal",
			"topic", topic)
		return
	default:
		var targetChan chan event
		switch priority {
		case HighPriority:
			targetChan = e.highch
		case MediumPriority:
			targetChan = e.mediumch
		case LowPriority:
			targetChan = e.lowch
		default:
			targetChan = e.lowch
		}

		select {
		case targetChan <- evt:
		default:
			slog.Error("Event channel is full, event sent to dead letter",
				"topic", topic,
				"priority", priority,
				"event_message", v,
				"channel_capacity", queueBufferSize,
			)
			deadLetterLogger.Error("Event channel full, event dropped",
				"topic", topic,
				"priority", priority,
				"event_message", v)
		}
	}
}
