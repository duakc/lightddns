package prometheus

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/duakc/lightddns/options"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewUsesDefaultPort(t *testing.T) {
	service, err := New(context.Background(), zap.NewNop(), options.PrometheusServiceOption{
		AbstractServiceOption: options.AbstractServiceOption{
			Type: ServiceType,
			Name: "test-prometheus",
		},
		Enabled: true,
	})
	require.NoError(t, err)
	require.Equal(t, ":9001", service.(*Prometheus).addr)
}

func TestCloseBeforeServeClosesListener(t *testing.T) {
	service := &Prometheus{
		server:    &http.Server{},
		listener:  closeTestListener{},
		serveErrC: make(chan error, 1),
	}

	done := make(chan error, 1)
	go func() { done <- service.Close() }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Close blocked before Serve started")
	}
}

type closeTestListener struct{}

func (closeTestListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (closeTestListener) Close() error              { return nil }
func (closeTestListener) Addr() net.Addr            { return closeTestAddr{} }

type closeTestAddr struct{}

func (closeTestAddr) Network() string { return "test" }
func (closeTestAddr) String() string  { return "test" }
