package event

// TODO: mutex is not needed currently
type DomainEventQueue struct {
	events []*DomainEvent
}

func (q *DomainEventQueue) DequeueEvent() *DomainEvent {
	if len(q.events) == 0 {
		return nil
	}

	first := q.events[0]
	q.events = q.events[1:]

	return first
}

func (q *DomainEventQueue) EnqueueEvent(e *DomainEvent) {
	q.events = append(q.events, e)
}
