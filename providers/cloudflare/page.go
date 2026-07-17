package cloudflare

import (
	"context"
	"io"
	"strconv"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"go.uber.org/zap"
)

type PageConfig[T any] struct {
	owner     *Client
	op        string
	reqConfig httpx.ReqConfig

	// internal
	page    int
	perPage int

	done bool

	resultInfo ResultInfo
}

func NewPaging[T any](owner *Client, op string, req httpx.ReqConfig) *PageConfig[T] {
	if owner == nil {
		panic("nil client")
	}
	return &PageConfig[T]{
		owner:     owner,
		op:        op,
		reqConfig: req,
		perPage:   50,
	}
}

func (pc *PageConfig[T]) Next(ctx context.Context) (_ []T, err error) {
	if pc.done {
		return nil, io.EOF
	}
	logger := pc.owner.actionLogger(pc.op).With(zap.Int("page", pc.page+1))
	logger.Debug("api call start")

	pc.page++
	pc.reqConfig.Query.Set("page", strconv.Itoa(pc.page))
	pc.reqConfig.Query.Set("per_page", strconv.Itoa(pc.perPage))

	result, response, perr := httpx.JSONRequest[ResponseWithPage[T]](ctx,
		pc.owner.do, pc.reqConfig, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if E := result.JoinError(perr); E != nil {
		logger.Warn("list page failed", zap.Error(E))
		err = E
		return nil, err
	}

	pc.resultInfo = result.ResultInfo
	pc.done = len(result.Result) == 0 || result.ResultInfo.Page >= result.ResultInfo.TotalPages
	return result.Result, nil
}

func (pc *PageConfig[T]) TotalCount() int {
	return pc.resultInfo.TotalCount
}
