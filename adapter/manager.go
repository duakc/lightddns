package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

var managerLogger = zaplog.NewPackage("adapter", "manager")

type ManagedType interface {
	Type() string
	Name() string
}

type AbstractManagedType struct {
	name string
	typ  string
}

func NewManagedType(typ, name string) AbstractManagedType {
	return AbstractManagedType{
		name: name,
		typ:  typ,
	}
}

func (a AbstractManagedType) Type() string {
	return a.typ
}

func (a AbstractManagedType) Name() string {
	return a.name
}

type Manager[T ManagedType] interface {
	services.ContextInjector

	Create(ctx context.Context, logger *zap.Logger, typ string, opt any) (T, error)
	Lookup(name string) (T, bool)
	LookupDefault(name string) (T, bool)
	LookupAll([]string) ([]T, error)
}

var _ Manager[ManagedType] = (*DefaultManager[ManagedType])(nil)

type DefaultManager[T ManagedType] struct {
	access sync.RWMutex

	items      []T
	itemByName map[string]T

	register Registry[T]
}

func (M *DefaultManager[T]) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[Manager[T]](ctx, M)
}

func NewManager[T ManagedType](R Registry[T]) *DefaultManager[T] {
	return &DefaultManager[T]{
		items:      make([]T, 0),
		itemByName: make(map[string]T),
		register:   R,
	}
}

func (M *DefaultManager[T]) Create(ctx context.Context, logger *zap.Logger, typ string, opt any) (T, error) {
	item, err := M.register.Create(ctx, logger, typ, opt)
	if err != nil {
		return mt.Zero[T](), err
	}
	if item.Type() == "" {
		return mt.Zero[T](), fmt.Errorf("empty type")
	}
	if item.Name() == "" {
		return mt.Zero[T](), fmt.Errorf("empty name")
	}

	M.access.Lock()
	defer M.access.Unlock()

	if _, existed := M.itemByName[item.Name()]; existed {
		return mt.Zero[T](), fmt.Errorf("duplicated: %q", item.Name())
	}

	M.items = append(M.items, item)
	M.itemByName[item.Name()] = item
	return item, nil
}

func (M *DefaultManager[T]) Lookup(name string) (T, bool) {
	M.access.RLock()
	defer M.access.RUnlock()

	if v, ok := M.itemByName[name]; ok {
		return v, true
	}
	return mt.Zero[T](), false
}

func (M *DefaultManager[T]) LookupDefault(name string) (T, bool) {
	M.access.RLock()
	defer M.access.RUnlock()

	if len(M.items) == 1 && name == "" {
		return M.items[0], true
	}

	if v, ok := M.itemByName[name]; ok {
		return v, true
	}

	return mt.Zero[T](), false
}

func (M *DefaultManager[T]) LookupAll(names []string) ([]T, error) {
	v := make([]T, 0, len(names))

	for _, name := range names {
		item, ok := M.LookupDefault(name)
		if !ok {
			return nil, NewManagedNotFoundError(name)
		}
		v = append(v, item)
	}
	return v, nil
}

//func CollectManagedItem[T ManagedType](M Manager[T], names []string) ([]T, error) {
//	var items []T
//	for _, name := range names {
//		item, ok := M.Lookup(name)
//		if !ok {
//			return nil, &ManagedNotFoundError{name}
//		}
//		items = append(items, item)
//	}
//	return items, nil
//}
