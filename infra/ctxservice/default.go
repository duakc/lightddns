package ctxservice

import "sync"

type DefaultRegistry struct {
	m *sync.Map
}

func NewDefaultRegistry() *DefaultRegistry {
	return &DefaultRegistry{m: &sync.Map{}}
}

func (d *DefaultRegistry) Store(k, v any) {
	d.m.Store(k, v)
}

func (d *DefaultRegistry) Load(k any) any {
	v, _ := d.m.Load(k)
	return v
}

func (d *DefaultRegistry) Clear() {
	d.m.Clear()
}
