package internal

import (
	"context"
	"io"
	"strconv"

	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/infra/zaplog"

	"go.uber.org/zap"
)

type PageConfig[T any] struct {
	reqConfig httpx.ReqConfig
	requester httpx.HTTPRequester

	// internal
	page    int
	perPage int

	done bool

	resultInfo ResultInfo
}

func NewPaging[T any](req httpx.ReqConfig, do httpx.HTTPRequester) *PageConfig[T] {
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
	logger := zaplog.FromOrPackage(ctx, "cloudflare", "internal").
		With(zap.String("action", "list_page"), zap.Int("page", pc.page+1))
	logger.Debug("cloudflare: api call start")
	pc.page++
	pc.reqConfig.Query.Set("page", strconv.Itoa(pc.page))
	pc.reqConfig.Query.Set("per_page", strconv.Itoa(pc.perPage))
	result, response, err := httpx.JSONRequest[ResponseWithPage[T]](ctx,
		pc.requester, pc.reqConfig, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}
	if E := result.JoinError(err); E != nil {
		logger.Warn("cloudflare: list page failed", zap.Error(E))
		return nil, E
	}

	pc.resultInfo = result.ResultInfo
	pc.done = result.ResultInfo.Page == result.ResultInfo.TotalPages
	return result.Result, nil
}

func (pc *PageConfig[T]) TotalCount() int {
	return pc.resultInfo.TotalCount
}
