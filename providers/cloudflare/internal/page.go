package internal

import (
	"context"
	"strconv"

	"github.com/duakc/lightddns/infra/httpxx"
)

type PageConfig[T any] struct {
	ReqConfig httpxx.ReqConfig
	Requester httpxx.HTTPRequester

	// internal
	page    int
	perPage int

	done bool
}

func NewPaging[T any](req httpxx.ReqConfig, do httpxx.HTTPRequester) *PageConfig[T] {
	if do == nil {
		panic("nil requester")
	}
	return &PageConfig[T]{
		ReqConfig: req,
		Requester: do,
		perPage:   50,
	}
}

func (pc *PageConfig[T]) Next(ctx context.Context) ([]T, error) {
	type responseType = apiResponse[T]
	pc.page++
	pc.ReqConfig.Query.Set("page", strconv.Itoa(pc.page))
	pc.ReqConfig.Query.Set("per_page", strconv.Itoa(pc.perPage))
	result, response, err := httpxx.JSONRequest[responseType](ctx,
		pc.Requester, pc.ReqConfig, nil)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	} else if err = result.Err(); err != nil {
		return nil, err
	}

	pc.done = result.ResultInfo.Page == result.ResultInfo.TotalPages
	return result.Results, nil
}
