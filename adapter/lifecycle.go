package adapter

import (
	"context"
	"fmt"
)

type Service interface {
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}

type PreStarter interface {
	PreStart(ctx context.Context) error
}

type PostStarter interface {
	PostStart(ctx context.Context) error
}

type PreCloser interface {
	PreClose(ctx context.Context) error
}
type PostCloser interface {
	PostClose(ctx context.Context) error
}

func StartService(ctx context.Context, ser any) error {
	var err error
	if preS, ok := ser.(PreStarter); ok {
		err = preS.PreStart(ctx)
		if err != nil {
			return newServiceLifeCycleError(err, "PreStart")
		}
	}
	if starter, ok := ser.(interface {
		Start(ctx context.Context) error
	}); ok {
		err = starter.Start(ctx)
		if err != nil {
			return newServiceLifeCycleError(err, "Start")
		}
	}
	if postS, ok := ser.(PostStarter); ok {
		err = postS.PostStart(ctx)
		if err != nil {
			return newServiceLifeCycleError(err, "PostStart")
		}
	}
	return nil
}

func CloseService(ctx context.Context, ser any) error {
	var err error
	if preS, ok := ser.(PreCloser); ok {
		err = preS.PreClose(ctx)
		if err != nil {
			return newServiceLifeCycleError(err, "PreClose")
		}
	}
	if starter, ok := ser.(interface {
		Close(ctx context.Context) error
	}); ok {
		err = starter.Close(ctx)
		if err != nil {
			return newServiceLifeCycleError(err, "Close")
		}
	}
	if postS, ok := ser.(PostCloser); ok {
		err = postS.PostClose(ctx)
		if err != nil {
			return newServiceLifeCycleError(err, "PostCose")
		}
	}
	return nil
}

type ServiceLifeCycleError struct {
	Err   error
	Stage string
}

func newServiceLifeCycleError(err error, stage string) error {
	return &ServiceLifeCycleError{
		Err:   err,
		Stage: stage,
	}
}

func (E *ServiceLifeCycleError) Error() string {
	return fmt.Sprintf("service stage: %s: %s", E.Stage, E.Err)
}

func (E *ServiceLifeCycleError) Unwrap() error {
	return E.Err
}
