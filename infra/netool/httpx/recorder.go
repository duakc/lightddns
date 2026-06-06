package httpx

import (
	"time"

	"github.com/duakc/mt/debug"

	"go.uber.org/zap"
)

// HTTPRequestRecorder logs the lifecycle of one outbound HTTP request. It's a
// debug aid for custom HTTPRequester implementations (e.g. a signing client's
// Do method) that want a uniform "started / ended / errored" trail without
// each provider reinventing the same zap.Debug pattern.
//
// Pure logging — metrics are recorded separately via ddnsmetric.ProviderAPIRouter.
type HTTPRequestRecorder struct {
	Logger *zap.Logger
	Err    *error

	Endpoint string

	Start time.Time
	End   time.Time
}

func NewHTTPRequestRecorder(logger *zap.Logger, endpoint string, err *error) *HTTPRequestRecorder {
	return &HTTPRequestRecorder{
		Logger:   logger,
		Err:      err,
		Endpoint: endpoint,
		Start:    time.Now(),
	}
}

func (r *HTTPRequestRecorder) Stop() {
	r.End = time.Now()
}

func (r *HTTPRequestRecorder) Record() {
	r.End = time.Now()
	consume := r.End.Sub(r.Start)

	var err error
	if r.Err != nil && *r.Err != nil {
		err = *r.Err
	}

	if debug.Enabled {
		r.Logger.Debug("api call record",
			zap.String("api_endpoint", r.Endpoint),
			zap.Duration("api_called_time", consume),
			zap.Bool("api_is_error", err != nil),
		)
	}

	if err == nil {
		return
	}

	if ce := r.Logger.Check(zap.ErrorLevel, "api call end"); ce != nil {
		fields := []zap.Field{
			zap.String("api_endpoint", r.Endpoint),
			zap.Duration("api_called_time", consume),
			zap.Error(err),
		}
		ce.Write(fields...)
	}
}
