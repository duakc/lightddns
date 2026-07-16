package transports

import (
	"crypto/tls"
	"fmt"

	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/internal"

	"go.uber.org/zap"
)

type TransportOptions struct {
	Logger *zap.Logger
	Type   string

	Dialer dialerx.Dialer

	Server     string
	ServerPort uint16
	TLSConfig  *tls.Config
}

func NewTransport(option TransportOptions) (Transport, error) {
	switch option.Type {
	case TransportTypeSystem:
		return NewSystemTransport(option.Logger, internal.DefaultDNSTTL), nil
	case TransportTypeTLS:
		return NewTLS(option.Logger, option.Dialer, option.Server, option.ServerPort, option.TLSConfig)
	default:
		return nil, fmt.Errorf("unknown transport type: %s", option.Type)
	}
}
