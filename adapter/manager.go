package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/duakc/lightddns/infra/common"
)

type managedType interface {
	Type() string
	Name() string
}

type Manager[T managedType] interface {
	Create(ctx context.Context, typ string, opt any) error
	Lookup(name string) (T, bool)
}

type DefaultManager[T managedType] struct {
	access sync.RWMutex

	items      []T
	itemByName map[string]T

	register Registry[T]
}

func NewManager[T managedType](R Registry[T]) *DefaultManager[T] {
	return &DefaultManager[T]{
		items:      make([]T, 0),
		itemByName: make(map[string]T),
		register:   R,
	}
}

func (M *DefaultManager[T]) Create(ctx context.Context, typ string, opt any) error {
	item, err := M.register.Create(ctx, typ, opt)
	if err != nil {
		return err
	}
	if item.Name() == "" {
		return fmt.Errorf("empty name")
	}
	if item.Type() == "" {
		return fmt.Errorf("empty type")
	}
	M.access.Lock()
	defer M.access.Unlock()
	if _, existed := M.itemByName[item.Name()]; existed {
		return fmt.Errorf("duplicated item: type=%s, name=%s", item.Type(), item.Name())
	}
	M.items = append(M.items, item)
	M.itemByName[item.Name()] = item
	return nil
}

func (M *DefaultManager[T]) Lookup(name string) (T, bool) {
	M.access.RLock()
	defer M.access.RUnlock()

	if v, ok := M.itemByName[name]; ok {
		return v, true
	}
	return common.Zero[T](), false
}
