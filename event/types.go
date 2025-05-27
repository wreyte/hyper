package event

type event struct {
	topic    string
	message  any
	priority Priority
}

type Subscription struct {
	Topic     string
	CreatedAt int64
	Fn        HandlerFunc
}
