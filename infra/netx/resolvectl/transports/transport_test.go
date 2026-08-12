package transports

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func TestWriteMessageFramesDNSOverTCP(t *testing.T) {
	message := testQueryMessage(1234)

	var written bytes.Buffer
	require.NoError(t, WriteMessage(&written, 42, message))

	frame := written.Bytes()
	require.GreaterOrEqual(t, len(frame), 2+12)
	require.Equal(t, len(frame)-2, int(binary.BigEndian.Uint16(frame[:2])))

	var unpacked dns.Msg
	require.NoError(t, unpacked.Unpack(frame[2:]))
	require.Equal(t, uint16(42), unpacked.Id)
	require.Equal(t, uint16(1234), message.Id)
}

func TestTLSTransportCreateSessionFailureDoesNotHoldLock(t *testing.T) {
	dialErr := errors.New("dial failed")
	transport, err := NewTLS(nil, errorDialer{err: dialErr}, "example.test", 853, &tls.Config{InsecureSkipVerify: true})
	require.NoError(t, err)

	_, err = transport.Exchange(context.Background(), testQueryMessage(1))
	require.ErrorIs(t, err, dialErr)

	done := make(chan error, 1)
	go func() {
		_, exchangeErr := transport.Exchange(context.Background(), testQueryMessage(2))
		done <- exchangeErr
	}()

	select {
	case err = <-done:
		require.ErrorIs(t, err, dialErr)
	case <-time.After(time.Second):
		t.Fatal("second exchange blocked after create session failure")
	}
}

func TestTLSTransportMarksClosedReusedSessionRetryable(t *testing.T) {
	dialer := &pipeTLSDialer{serverConfig: testServerTLSConfig(t)}
	transport, err := NewTLS(nil, dialer, "example.test", 853, &tls.Config{InsecureSkipVerify: true})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, transport.Close())
	})

	response, err := transport.Exchange(context.Background(), testQueryMessage(100))
	require.NoError(t, err)
	require.Equal(t, uint16(100), response.Id)
	require.Len(t, response.Answer, 1)

	_, err = transport.Exchange(context.Background(), testQueryMessage(101))
	require.Error(t, err)
	require.True(t, IsRetryable(err))
	require.Equal(t, int32(1), dialer.dials.Load())
}

func TestTLSTransportReturnsEOFOnFreshSession(t *testing.T) {
	dialer := &closingTLSDialer{serverConfig: testServerTLSConfig(t)}
	transport, err := NewTLS(nil, dialer, "example.test", 853, &tls.Config{InsecureSkipVerify: true})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, transport.Close())
	})

	_, err = transport.Exchange(context.Background(), testQueryMessage(102))
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Equal(t, int32(1), dialer.dials.Load())
}

type errorDialer struct {
	err error
}

func (d errorDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	return nil, d.err
}

type pipeTLSDialer struct {
	serverConfig *tls.Config
	dials        atomic.Int32
}

func (d *pipeTLSDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	clientConn, serverConn := net.Pipe()
	d.dials.Add(1)

	serverTLS := tls.Server(serverConn, d.serverConfig.Clone())
	go func() {
		defer serverConn.Close()
		if err := serverTLS.Handshake(); err != nil {
			return
		}
		request, err := ReadMessage(serverTLS)
		if err != nil {
			return
		}
		response := new(dns.Msg)
		response.SetReply(request)
		response.Answer = []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   request.Question[0].Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: net.IPv4(192, 0, 2, 1),
			},
		}
		_ = WriteMessage(serverTLS, request.Id, response)
	}()

	return clientConn, nil
}

type closingTLSDialer struct {
	serverConfig *tls.Config
	dials        atomic.Int32
}

func (d *closingTLSDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	clientConn, serverConn := net.Pipe()
	d.dials.Add(1)

	serverTLS := tls.Server(serverConn, d.serverConfig.Clone())
	go func() {
		defer func() {
			_ = serverConn.Close()
		}()
		if err := serverTLS.Handshake(); err != nil {
			return
		}
		_ = serverConn.Close()
	}()

	return clientConn, nil
}

func testQueryMessage(id uint16) *dns.Msg {
	message := new(dns.Msg)
	message.SetQuestion("example.test.", dns.TypeA)
	message.Id = id
	return message
}

func testServerTLSConfig(t *testing.T) *tls.Config {
	t.Helper()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "example.test",
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(time.Hour),
		DNSNames:  []string{"example.test"},
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	certificate, err := tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
	)
	require.NoError(t, err)

	return &tls.Config{Certificates: []tls.Certificate{certificate}}
}
