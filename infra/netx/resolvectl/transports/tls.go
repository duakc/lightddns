package transports

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/duakc/lightddns/infra/netx/dialerx"

	"github.com/duakc/mt/common/generic"

	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
)

const TransportTypeTLS = "tls"

type tlsTransportConn struct {
	conn    *tls.Conn
	queryID uint16
}

type TLSTransport struct {
	logger *zap.Logger

	dialer    dialerx.Dialer
	server    string
	tlsConfig *tls.Config

	access      sync.Mutex
	tlsSessions *generic.List[*tlsTransportConn]
	closed      atomic.Bool
}

func NewTLS(logger *zap.Logger, dialer dialerx.Dialer, server string, serverPort uint16, tlsConfig *tls.Config) (*TLSTransport, error) {
	if serverPort == 0 {
		serverPort = 853
	}
	if server == "" {
		return nil, errors.New("tls: empty server address")
	}
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	return &TLSTransport{
		logger:      createLogger(logger, TransportTypeTLS),
		dialer:      dialer,
		server:      net.JoinHostPort(server, strconv.Itoa(int(serverPort))),
		tlsConfig:   tlsConfig,
		tlsSessions: generic.NewList[*tlsTransportConn](),
	}, nil
}

func (t *TLSTransport) Exchange(ctx context.Context, message *mDns.Msg) (*mDns.Msg, error) {
	if t.closed.Load() {
		return nil, os.ErrClosed
	}
	t.access.Lock()
	tlsSession, ok := t.tlsSessions.PopFront()
	if !ok {
		var err error
		tlsSession, err = t.createSession(ctx)
		if err != nil {
			return nil, fmt.Errorf("tls: create tls session: %w", err)
		}
	}
	t.access.Unlock()
	if deadline, ok := ctx.Deadline(); ok {
		_ = tlsSession.conn.SetDeadline(deadline)
	}

	messageID := message.Id
	tlsSession.queryID++
	if err := WriteMessage(tlsSession.conn, tlsSession.queryID, message); err != nil {
		_ = tlsSession.conn.Close()
		return nil, fmt.Errorf("tls: write: %w", err)
	}
	exchangedMessage, err := ReadMessage(tlsSession.conn)
	if err != nil {
		_ = tlsSession.conn.Close()
		return nil, fmt.Errorf("tls: read: %w", err)
	}
	exchangedMessage.Id = messageID
	_ = tlsSession.conn.SetDeadline(time.Time{})
	if !t.closed.Load() {
		t.access.Lock()
		t.tlsSessions.PushBack(tlsSession)
		t.access.Unlock()
	} else {
		_ = tlsSession.conn.Close()
	}
	return exchangedMessage, nil
}

func (t *TLSTransport) createSession(ctx context.Context) (*tlsTransportConn, error) {
	conn, err := t.dialer.DialContext(ctx, "tcp", t.server)
	if err != nil {
		return nil, err
	}
	tlsConn := tls.Client(conn, t.tlsConfig)
	err = tlsConn.HandshakeContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("tls handshake: %w", err)
	}
	return &tlsTransportConn{tlsConn, 0}, nil
}

func (t *TLSTransport) Close() error {
	t.closed.Store(true)
	t.access.Lock()
	var next *generic.ListElement[*tlsTransportConn]
	for head := t.tlsSessions.Front(); head != nil; head = next {
		_ = head.Value.conn.Close()
		next = head.Next()
		t.tlsSessions.Remove(head)
	}
	t.access.Unlock()
	return nil
}
