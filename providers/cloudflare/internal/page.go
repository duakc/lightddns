package internal

import (
	"context"
	"io"
	"strconv"

	"github.com/duakc/lightddns/infra/httpxx"
)

type PageConfig[T any] struct {
	reqConfig httpxx.ReqConfig
	requester httpxx.HTTPRequester

	// internal
	page    int
	perPage int

	done bool

	resultInfo ResultInfo
}

func NewPaging[T any](req httpxx.ReqConfig, do httpxx.HTTPRequester) *PageConfig[T] {
	if do == nil {
		panic("nil requester")
	}
	return &PageConfig[T]{
		reqConfig: req,
		requester: do,
		perPage:   50,
	}
}

func (pc *PageConfig[T]) Next(ctx context.Context) ([]T, error) {
	if pc.done {
		return nil, io.EOF
	}
	pc.page++
	pc.reqConfig.Query.Set("page", strconv.Itoa(pc.page))
	pc.reqConfig.Query.Set("per_page", strconv.Itoa(pc.perPage))
	result, response, err := httpxx.JSONRequest[ResponseWithPage[T]](ctx,
		pc.requester, pc.reqConfig, nil)
	if response != nil {
		defer response.Body.Close()
	}
	if E := result.JoinError(err); E != nil {
		return nil, E
	}
	
	pc.resultInfo = result.ResultInfo
	pc.done = result.ResultInfo.Page == result.ResultInfo.TotalPages
	return result.Results, nil
}

func (pc *PageConfig[T]) TotalCount() int {
	return pc.resultInfo.TotalCount
}
