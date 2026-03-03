package audit

import (
	"sync"
	"testing"
	"time"
)

type mockObserver struct {
	mu     sync.Mutex
	events []Event
}

func (m *mockObserver) Notify(e Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
}

func (m *mockObserver) getEvents() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]Event, len(m.events))
	copy(cp, m.events)
	return cp
}

func TestNewPublisher(t *testing.T) {
	p := NewPublisher()
	if p == nil {
		t.Fatal("NewPublisher вернул nil")
	}
	if len(p.observers) != 0 {
		t.Errorf("ожидали 0 observers, получили %d", len(p.observers))
	}
}

func TestSubscribe(t *testing.T) {
	p := NewPublisher()
	o1 := &mockObserver{}
	o2 := &mockObserver{}

	p.Subscribe(o1)
	if len(p.observers) != 1 {
		t.Errorf("ожидали 1 observer, получили %d", len(p.observers))
	}

	p.Subscribe(o2)
	if len(p.observers) != 2 {
		t.Errorf("ожидали 2 observers, получили %d", len(p.observers))
	}
}

func TestPublish(t *testing.T) {
	p := NewPublisher()
	o1 := &mockObserver{}
	o2 := &mockObserver{}
	p.Subscribe(o1)
	p.Subscribe(o2)

	event := Event{
		TS:     1000,
		Action: "shorten",
		UserID: "user1",
		URL:    "https://example.com",
	}
	p.Publish(event)

	// Publish вызывает Notify в горутинах, ждём
	time.Sleep(100 * time.Millisecond)

	events1 := o1.getEvents()
	events2 := o2.getEvents()

	if len(events1) != 1 {
		t.Errorf("observer1: ожидали 1 событие, получили %d", len(events1))
	}
	if len(events2) != 1 {
		t.Errorf("observer2: ожидали 1 событие, получили %d", len(events2))
	}

	if len(events1) > 0 && events1[0].Action != "shorten" {
		t.Errorf("ожидали action=shorten, получили %s", events1[0].Action)
	}
}

func TestPublishNoObservers(t *testing.T) {
	p := NewPublisher()
	// не должен паниковать
	p.Publish(Event{Action: "test"})
}
