package audit

type Publisher struct {
	observers []Observer
}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) Subscribe(o Observer) {
	p.observers = append(p.observers, o)
}

func (p *Publisher) Publish(e Event) {
	for _, o := range p.observers {
		go o.Notify(e)
	}
}
