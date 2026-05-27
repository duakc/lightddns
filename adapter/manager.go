package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services"
)

// use a optionalKey to avoid use a special value of string
// it use a isEmpty indicated that the value is empty.
type optionalKey[T any] struct {
	isEmpty bool
	v       T
}

func newOptionalKey[T comparable](key T) optionalKey[T] {
	if key == mt.Zero[T]() {
		return optionalKey[T]{isEmpty: true}
	}
	return optionalKey[T]{v: key}
}

var managerLogger = zaplog.NewPackage("adapter", "manager")

type managedType interface {
	Type() string
	Name() string
}

type AbstractManagedType struct {
	name string
	typ  string
}

func NewManagedType(typ, name string) AbstractManagedType {
	return AbstractManagedType{name, typ}
}

func (a AbstractManagedType) Type() string {
	return a.typ
}

func (a AbstractManagedType) Name() string {
	return a.name
}

type Manager[T managedType] interface {
	services.ContextInjector

	Create(ctx context.Context, typ string, opt any) error
	Lookup(name string) (T, bool)
}

var _ Manager[managedType] = (*DefaultManager[managedType])(nil)

type DefaultManager[T managedType] struct {
	access sync.RWMutex

	items      []T
	itemByName map[optionalKey[string]]T

	register Registry[T]

	elementContainEmptyKey bool
}

func (M *DefaultManager[T]) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[Manager[T]](ctx, M)
}

func NewManager[T managedType](R Registry[T]) *DefaultManager[T] {
	return &DefaultManager[T]{
		items:      make([]T, 0),
		itemByName: make(map[optionalKey[string]]T),
		register:   R,
	}
}

func (M *DefaultManager[T]) Create(ctx context.Context, typ string, opt any) error {
	item, err := M.register.Create(ctx, typ, opt)
	if err != nil {
		return err
	}

	if item.Type() == "" {
		return fmt.Errorf("empty type")
	}
	name := newOptionalKey(item.Name())

	// in short: we only allow one elem's name is empty.
	// otherwise, return an error.
	if name.isEmpty && len(M.items) == 0 {
		// this is the first element to this manager
		// we allow empty name for this.
		M.access.Lock()
		M.elementContainEmptyKey = true
	} else if name.isEmpty {
		// if already had an elem in this manager but the `name` is also empty,
		// we return an error.
		return fmt.Errorf("empty name")
	} else {
		M.access.Lock()
	}
	defer M.access.Unlock()

	if _, existed := M.itemByName[name]; existed {
		return fmt.Errorf("duplicated: %q", item.Name())
	}
	M.items = append(M.items, item)
	M.itemByName[name] = item
	return nil
}

func (M *DefaultManager[T]) Lookup(name string) (T, bool) {
	M.access.RLock()
	defer M.access.RUnlock()
	k := newOptionalKey(name)

	if v, ok := M.itemByName[k]; ok {
		return v, true
	}
	return mt.Zero[T](), false
}

func CollectManagedItem[T managedType](M Manager[T], names []string) ([]T, error) {
	var items []T
	for _, name := range names {
		item, ok := M.Lookup(name)
		if !ok {
			return nil, &ManagedNotFoundError{name}
		}
		items = append(items, item)
	}
	return items, nil
}
