package pool

import "sync"

// Resetter — ограничение для типов, которые можно класть в пул
type Resetter interface {
	Reset()
}

// Pool — generic-контейнер для объектов с Reset()
type Pool[T Resetter] struct {
	pool *sync.Pool
}

// New — конструктор пула
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get — получить объект из пула
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put — сбросить состояние и вернуть объект в пул
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
