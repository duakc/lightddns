package httpx

import (
	"net/http"
	"time"

	"github.com/duakc/mt/debug"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
	if debug.Enabled {
		logger = logger.With(zap.String("http_request_id", uuid.NewString()))
	}

	logger.Debug("http record start")

	return &HTTPRequestRecorder{
		Request:  req,
		Response: resp,
		Logger:   logger,
		Err:      err,

		Start: time.Now(),
	}
}

func (r *HTTPRequestRecorder) Stop() {
	r.End = time.Now()
}

func (r *HTTPRequestRecorder) Record() {
	if r.End.IsZero() {
		r.End = time.Now()
	}

	consume := r.End.Sub(r.Start)

	var err error
	if r.Err != nil && *r.Err != nil {
		err = *r.Err
	}

	if err == nil && !debug.Enabled {
		r.Logger.Debug("http record end",
			zap.Duration("http_request_duration", consume))
		return
	}

	checkLevel := zapcore.WarnLevel
	if debug.Enabled {
		checkLevel = zapcore.DebugLevel
	}

	if ce := r.Logger.Check(checkLevel, "http record end"); ce != nil {
		fields := []zap.Field{
			zap.Duration("http_request_duration", consume),
		}
		if r.Request != nil {
			if debug.Enabled {
				fields = append(fields,
					zap.String("http_request_query", r.Request.URL.RawQuery),
				)
			}
			fields = append(fields,
				zap.String("http_request_path", r.Request.URL.Path))
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		if r.Response != nil && *r.Response != nil {
			fields = append(fields,
				zap.Int("http_response_code", (*r.Response).StatusCode),
			)
			if debug.Enabled {
				fields = append(fields,
					zap.Any("http_response_header", (*r.Response).Header),
				)
			}
		}
		ce.Write(fields...)
	}
}
