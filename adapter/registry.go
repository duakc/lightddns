package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/duakc/mt"
	"github.com/duakc/mt/debug"

	"go.uber.org/zap"
)

func Register[T managedType, O any](R Registry[T], typ string, constructor GenericObjectConstructor[T, O]) {
	managerLogger.Trace("new type registered", zap.String("type", typ))
	R.registry(typ, func() any {
		return new(O)
	}, func(ctx context.Context, option any) (T, error) {
		var opt *O
		if option != nil {
			opt = option.(*O)
		}
		return constructor(ctx, mt.PtrValueOrDefault(opt))
	})
}

type Registry[T managedType] interface {
	Create(ctx context.Context, typ string, option any) (T, error)
	CreateOption(typ string) (any, error)

	registry(typ string, option optionConstructor, object objectConstructor[T])
}

type (
	GenericObjectConstructor[T managedType, O any] func(ctx context.Context, option O) (T, error)

	optionConstructor                func() any
	objectConstructor[T managedType] func(ctx context.Context, option any) (T, error)
)

func NewRegister[T managedType]() Registry[T] {
	return &defaultRegistry[T]{
		typeToObject: make(map[string]objectConstructor[T]),
		typeToOption: make(map[string]optionConstructor),
	}
}

type defaultRegistry[T managedType] struct {
	access sync.Mutex

	typeToOption map[string]optionConstructor
	typeToObject map[string]objectConstructor[T]
}

func (R *defaultRegistry[T]) Create(ctx context.Context, typ string, option any) (T, error) {
	R.access.Lock()
	defer R.access.Unlock()

	create := R.typeToObject[typ]
	if create == nil {
		return mt.Zero[T](), fmt.Errorf("unregistered type: %s", typ)
	}

	returned, err := create(ctx, option)
	if err != nil {
		return mt.Zero[T](), err
	}
	if err == nil && debug.Enabled && returned.Type() != typ {
		panic("mismatch type excepted: " + typ + ", got: " + returned.Type())
	}

	managerLogger.Trace("registered object created",
		zap.String("type", typ),
		zap.String("name", returned.Name()))

	return returned, nil
}

func (R *defaultRegistry[T]) CreateOption(typ string) (any, error) {
	R.access.Lock()
	defer R.access.Unlock()

	create := R.typeToOption[typ]
	if create == nil {
		return nil, fmt.Errorf("unregistered type: %s", typ)
	}
	managerLogger.Trace("new option created", zap.String("type", typ))
	return create(), nil
}

func (R *defaultRegistry[T]) registry(typ string, option optionConstructor, object objectConstructor[T]) {
	R.access.Lock()
	defer R.access.Unlock()

	R.typeToOption[typ] = option
	R.typeToObject[typ] = object
}
