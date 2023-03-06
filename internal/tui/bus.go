package tui

type EventBus struct {
	subscribers map[string][]EventBusEventCallback
}

type EventBusEventCallback func(data interface{})

func (bus *EventBus) Publish(name string, data interface{}) {
	for _, v := range bus.subscribers[name] {
		v(data)
	}
}

func (bus *EventBus) Subscribe(name string, callback EventBusEventCallback) {
	if eventBus.subscribers[name] == nil {
		eventBus.subscribers[name] = make([]EventBusEventCallback, 0)
	}

	eventBus.subscribers[name] = append(eventBus.subscribers[name], callback)
}

func NewEventBus() *EventBus {
	subscribers := make(map[string][]EventBusEventCallback)
	return &EventBus{
		subscribers: subscribers,
	}
}
