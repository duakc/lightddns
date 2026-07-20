package httpx

import (
	"net/http"
	"time"

	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var defaultLogger = zaplog.NewPackage("netx", "httpx")

type HTTPRequestRecorder struct {
	Request  *http.Request
	Response **http.Response

	Logger *zap.Logger
	Err    *error

	Start time.Time
	End   time.Time

	// debug only
	id string
}

func NewHTTPRequestRecorder(logger *zap.Logger, req *http.Request,
	resp **http.Response, err *error,
) *HTTPRequestRecorder {
	requestId := uuid.New().String()
	logger = logger.WithLazy(
		zap.String("http_request_id", requestId),
		zap.String("http_request_method", req.Method),
		zap.String("http_request_url", req.URL.String()),
		zap.Any("http_request_headers", req.Header),
		zap.Any("http_request_body", req.Body),
	)

	logger.Debug("http record start")

	return &HTTPRequestRecorder{
		Request:  req,
		Response: resp,
		Logger:   logger,
		Err:      err,

		Start: time.Now(),

		id: requestId,
	}
}

func (r *HTTPRequestRecorder) Stop() {
	if r.End.IsZero() {
		r.End = time.Now()
	}
}

func (r *HTTPRequestRecorder) Record() {
	r.Stop()

	if !r.Logger.Level().Enabled(zap.DebugLevel) {
		return
	}

	consume := r.End.Sub(r.Start)

	fields := []zap.Field{
		zap.Duration("http_response_duration", consume),
	}
	if r.Err != nil && *r.Err != nil {
		fields = append(fields, zap.Error(*r.Err))
	}

	if r.Response != nil && *r.Response != nil {
		fields = append(fields,
			zap.Any("http_response_header", (*r.Response).Header),
			zap.Any("http_response_body", (*r.Response).Body),
			zap.Int("http_response_status_code", (*r.Response).StatusCode),
		)
	}
	r.Logger.Debug("http record end", fields...)
}
