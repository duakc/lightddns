package prometheus

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/metrics"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const ServiceType = constpkg.ServiceTypePrometheus

const (
	defaultPort = 9090
	defaultPath = "/metrics"
)

func init() {
	adapter.Register(
		adapter.ServiceRegistry,
		ServiceType,
		New,
	)
}

type Prometheus struct {
	adapter.AbstractManagedType

	logger *zap.Logger
	addr   string
	path   string

	server    *http.Server
	listener  net.Listener
	serveErrC chan error
}

func New(ctx context.Context, option options.PrometheusServiceOption) (adapter.Service, error) {
	if !option.Enabled {
		return nil, adapter.ErrManagedItemNotEnabled
	}
	port := option.Port
	if port == 0 {
		port = defaultPort
	}
	httpPath := option.Path
	if httpPath == "" {
		httpPath = defaultPath
	}

	prometheus := &Prometheus{
		AbstractManagedType: adapter.NewManagedType(ServiceType, option.Name),
		addr:                net.JoinHostPort(option.Listen, strconv.FormatUint(uint64(port), 10)),
		path:                path.Clean(httpPath),
	}
	prometheus.logger = adapter.CreateServiceLogger(zaplog.FromContext(ctx), prometheus)
	return prometheus, nil
}

func (s *Prometheus) Start(ctx context.Context, stage services.Stage) error {
	switch stage {
	case services.StagePreStart:
		registry := services.Lookup[metrics.Registry](ctx)
		if registry == nil {
			return errors.New("metrics registry not found in context")
		}
		mux := http.NewServeMux()
		mux.Handle(s.path, promhttp.HandlerFor(registry.Gatherer(), promhttp.HandlerOpts{}))
		s.server = &http.Server{
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
		if s.server == nil || s.listener == nil {
			return errors.New("prometheus service not pre-started")
		}
		s.logger.Info("prometheus exporter listening",
			zap.String("addr", s.listener.Addr().String()),
			zap.String("path", s.path))
		go func() {
			err := s.server.Serve(s.listener)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.Error("prometheus server exited", zap.Error(err))
				s.serveErrC <- err
			}
			close(s.serveErrC)
		}()
		return nil
	case services.StagePostStart:
		return nil
	default:
		panic("unknown stage: " + stage.String())
	}
}

func (s *Prometheus) Close() error {
	if s.server == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(shutdownCtx)
	if serveErr, ok := <-s.serveErrC; ok {
		err = errors.Join(err, serveErr)
	}
	return err
}
