package httpx

import (
	"net/http"
	urlpkg "net/url"
	"strings"
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

const redactedValue = "[REDACTED]"

func RedactQuery(query urlpkg.Values) urlpkg.Values {
	redacted := make(urlpkg.Values, len(query))
	for key, values := range query {
		if isSensitiveName(key) {
			redacted.Set(key, redactedValue)
			continue
		}
		redacted[key] = append([]string(nil), values...)
	}
	return redacted
}

func RedactHeader(header http.Header) http.Header {
	redacted := header.Clone()
	for key := range redacted {
		if isSensitiveName(key) {
			redacted.Set(key, redactedValue)
		}
	}
	return redacted
}

func isSensitiveName(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "authorization") ||
		strings.Contains(name, "accesskey") ||
		strings.Contains(name, "api-key") ||
		strings.Contains(name, "apikey") ||
		strings.Contains(name, "credential") ||
		strings.Contains(name, "secret") ||
		strings.Contains(name, "signature") ||
		strings.Contains(name, "token")
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
					zap.String("http_request_query", RedactQuery(r.Request.URL.Query()).Encode()),
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
					zap.Any("http_response_header", RedactHeader((*r.Response).Header)),
				)
			}
		}
		ce.Write(fields...)
	}
}
