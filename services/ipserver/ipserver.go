package ipserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/freebuf"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RespFormat string

const (
	FormatEmpty RespFormat = ""
	FormatJson  RespFormat = "json"
	FormatYaml  RespFormat = "yaml"
)

var _ adapter.Service = (*IPServer)(nil)

const (
	ServiceType = constpkg.ServiceTypeIPServer
	DefaultPath = "/"
)

func init() {
	adapter.Register(
		adapter.ServiceRegistry,
		ServiceType,
		New,
	)
}

func New(ctx context.Context, logger *zap.Logger, option options.IPServerServiceOption) (adapter.Service, error) {
	if !option.Enabled {
		return nil, adapter.ErrManagedItemNotEnabled
	}

	if option.Port == 0 {
		return nil, fmt.Errorf("missing port")
	}

	if option.Path == "" {
		option.Path = "/"
	}
	return &IPServer{
		AbstractManagedType: adapter.NewManagedType(option.Type, option.Name),
		logger:              logger,
		addr: net.JoinHostPort(option.Listen,
			strconv.FormatUint(uint64(option.Port), 10)),
		path: option.Path,
		dump: option.Dump,
	}, nil
}

type IPServer struct {
	adapter.AbstractManagedType

	logger *zap.Logger

	addr string
	path string
	dump bool

	listener   net.Listener
	httpserver *http.Server

	serveErrC chan error
}

func (s *IPServer) Start(ctx context.Context, stage services.Stage) error {
	switch stage {
	case services.StagePreStart:
		mux := http.NewServeMux()
		mux.Handle(s.path, s)
		s.httpserver = &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}
		listener, err := net.Listen("tcp", s.addr)
		if err != nil {
			return fmt.Errorf("listen %s: %w", s.addr, err)
		}
		s.listener = listener
		s.serveErrC = make(chan error, 1)
		return nil
	case services.StageStart:
		if s.httpserver == nil || s.listener == nil {
			return errors.New("prometheus service not pre-started")
		}
		s.logger.Info("ip server started",
			zap.String("addr", s.listener.Addr().String()),
			zap.String("path", s.path))
		go func() {
			err := s.httpserver.Serve(s.listener)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.Error("ip server exited", zap.Error(err))
				s.serveErrC <- err
			}
			close(s.serveErrC)
		}()
		return nil
	case services.StagePostStart:
	default:
		panic(fmt.Sprintf("unknown stage: %d", stage))
	}
	return nil
}

func (s *IPServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger := s.logger

	if ce := logger.Check(zap.DebugLevel, "dump"); s.dump && ce != nil {
		loggingResponse := &loggingResponseWriter{
			ResponseWriter: resp,
			req:            req,
			body:           &bytes.Buffer{},
		}
		resp = loggingResponse

		defer loggingResponse.logging(ce)
	}

	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		resp.Header().Set("Allow", http.MethodGet)
		return
	}

	realIPStr, _, netipAddr := extractIP(req)
	if len(realIPStr) == 0 {
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte("ip not found in request"))
		return
	}

	var err error
	urlQueryFormat := strings.ToLower(req.URL.Query().Get("format"))
	switch urlQueryFormat {
	case string(FormatEmpty):
		resp.WriteHeader(http.StatusOK)
		_, err = resp.Write(unsafe.Slice(unsafe.StringData(realIPStr), len(realIPStr)))
	case string(FormatJson):
		resp.WriteHeader(http.StatusOK)
		err = s.serverHTTPJson(resp, req, realIPStr, netipAddr)
	case string(FormatYaml):
		resp.WriteHeader(http.StatusOK)
		err = s.serverHTTPYaml(resp, req, realIPStr, netipAddr)
	default:
		resp.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(resp, "unknown format: %s", urlQueryFormat)
		return
	}

	if err != nil {
		logger.Info("failed to write response", zap.Error(err))
	}
}

func (s *IPServer) serverHTTPJson(resp http.ResponseWriter, req *http.Request, ip string, netipAddr netip.Addr) error {
	respObj := &Response{
		IP:      ip,
		IsBogon: netool.IsBogon(netipAddr),
	}
	buffer := freebuf.NewSerial()
	respObj.writeJSON(buffer)
	_, err := buffer.WriteTo(resp)
	buffer.FreeMe()
	return err
}

func (s *IPServer) serverHTTPYaml(resp http.ResponseWriter, req *http.Request, ip string, netipAddr netip.Addr) error {
	respObj := &Response{
		IP:      ip,
		IsBogon: netool.IsBogon(netipAddr),
	}
	buffer := freebuf.NewSerial()
	respObj.writeYAML(buffer)
	_, err := buffer.WriteTo(resp)
	buffer.FreeMe()
	return err
}

// reimplement of httpx.ExtractIPFromRequest
func extractIP(req *http.Request) (string, string, netip.Addr) {
	for _, header := range []string{
		"Cf-Connecting-IP",
		"True-Client-IP",
		"X-Real-IP",
		"X-Forwarded-For",
	} {
		currentHeader := req.Header.Get(header)
		if currentHeader == "" {
			continue
		}
		cutHeader, _, found := strings.Cut(currentHeader, ",")
		if !found {
			cutHeader = currentHeader
		}
		ipStr := strings.TrimSpace(cutHeader)
		if ipStr == "" {
			continue
		}
		if addr, err := netip.ParseAddr(ipStr); err == nil {
			return ipStr, currentHeader, addr
		}
	}

	// see bench_test.go
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", "", netip.Addr{}
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return "", "", netip.Addr{}
	}
	//

	return host, "", addr
}

func (s *IPServer) Close() error {
	if s.httpserver == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.httpserver.Shutdown(shutdownCtx)
	if serveErr, ok := <-s.serveErrC; ok {
		err = errors.Join(err, serveErr)
	}
	return err
}

type loggingResponseWriter struct {
	http.ResponseWriter

	code int
	req  *http.Request
	body *bytes.Buffer
}

func (w *loggingResponseWriter) logging(ce *zapcore.CheckedEntry) {
	if w == nil || ce == nil {
		return
	}

	ce.Write(
		// request
		zap.String("request_method", w.req.Method),
		zap.String("request_uri", w.req.RequestURI),
		zap.Any("request_headers", w.req.Header),

		// response
		zap.Int("response_code", w.code),
		zap.Any("response_header", w.Header()),
		zap.String("response_body", w.body.String()),
	)
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *loggingResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
