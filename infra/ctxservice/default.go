package ctxservice

import "sync"

type defaultRegistry struct {
	m *sync.Map
}

func newDefaultRegistry() *defaultRegistry {
	return &defaultRegistry{m: &sync.Map{}}
}

func (d *defaultRegistry) Store(k, v any) {
	d.m.Store(k, v)
}

func (d *defaultRegistry) Load(k any) any {
	v, _ := d.m.Load(k)
	return v
}

func (d *defaultRegistry) Clear() {
	d.m.Clear()
}
