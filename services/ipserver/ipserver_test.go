package ipserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duakc/lightddns/options"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewUsesDefaultPort(t *testing.T) {
	service, err := New(context.Background(), zap.NewNop(), options.IPServerServiceOption{
		AbstractServiceOption: options.AbstractServiceOption{
			Type: ServiceType,
			Name: "test-ipserver",
		},
		Enabled: true,
	})
	require.NoError(t, err)
	require.Equal(t, ":9002", service.(*IPServer).addr)
}

func newTestIPServer(dump bool) *IPServer {
	return &IPServer{
		logger: zap.NewNop(),
		path:   DefaultPath,
		dump:   dump,
	}
}

func doServe(t *testing.T, s *IPServer, target string, remoteAddr string, headers http.Header) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	} else {
		req.RemoteAddr = ""
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	return rec
}

func TestServeHTTP_PlainTextFromRemoteAddr(t *testing.T) {
	s := newTestIPServer(false)
	rec := doServe(t, s, "/", "203.0.113.5:54321", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "203.0.113.5", rec.Body.String())
}

func TestServeHTTP_IPv6FromRemoteAddr(t *testing.T) {
	s := newTestIPServer(false)
	rec := doServe(t, s, "/", "[2001:db8::1]:443", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "2001:db8::1", rec.Body.String())
}

func TestServeHTTP_InvalidRemoteAddr(t *testing.T) {
	s := newTestIPServer(false)
	rec := doServe(t, s, "/", "not-an-addr", nil)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "ip not found")
}

func TestServeHTTP_HeaderPriority(t *testing.T) {
	cases := []struct {
		name    string
		headers http.Header
		want    string
	}{
		{
			name: "X-Forwarded-For only",
			headers: http.Header{
				"X-Forwarded-For": []string{"198.51.100.10, 10.0.0.1"},
			},
			want: "198.51.100.10",
		},
		{
			name: "X-Real-IP wins over XFF",
			headers: http.Header{
				"X-Real-Ip":       []string{"198.51.100.20"},
				"X-Forwarded-For": []string{"198.51.100.10"},
			},
			want: "198.51.100.20",
		},
		{
			name: "True-Client-IP wins over X-Real-IP",
			headers: http.Header{
				"True-Client-Ip": []string{"198.51.100.30"},
				"X-Real-Ip":      []string{"198.51.100.20"},
			},
			want: "198.51.100.30",
		},
		{
			name: "Cf-Connecting-IP wins over all",
			headers: http.Header{
				"Cf-Connecting-Ip": []string{"198.51.100.40"},
				"True-Client-Ip":   []string{"198.51.100.30"},
				"X-Real-Ip":        []string{"198.51.100.20"},
				"X-Forwarded-For":  []string{"198.51.100.10"},
			},
			want: "198.51.100.40",
		},
		{
			name: "IPv6 via header",
			headers: http.Header{
				"X-Real-Ip": []string{"2001:db8::abcd"},
			},
			want: "2001:db8::abcd",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestIPServer(false)
			rec := doServe(t, s, "/", "203.0.113.99:1234", tc.headers)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tc.want, rec.Body.String())
		})
	}
}

func TestServeHTTP_HeaderInvalidIPsFallback(t *testing.T) {
	s := newTestIPServer(false)
	headers := http.Header{
		"X-Forwarded-For": []string{"not-an-ip, also-bad"},
	}
	rec := doServe(t, s, "/", "203.0.113.7:80", headers)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "203.0.113.7", rec.Body.String())
}

func TestServeHTTP_FormatJSON(t *testing.T) {
	s := newTestIPServer(false)
	rec := doServe(t, s, "/?format=json", "203.0.113.5:1234", nil)

	require.Equal(t, http.StatusOK, rec.Code)

	var got Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got),
		"response body must be valid JSON: %s", rec.Body.String())

	assert.Equal(t, "203.0.113.5", got.IP)
	assert.False(t, got.IsBogon, "TEST-NET-3 is publicly-routable per netx.IsBogon")
}

func TestServeHTTP_FormatJSONCaseInsensitive(t *testing.T) {
	s := newTestIPServer(false)
	rec := doServe(t, s, "/?format=JSON", "203.0.113.5:1234", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var got Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "203.0.113.5", got.IP)
}

func TestServeHTTP_FormatYAML(t *testing.T) {
	s := newTestIPServer(false)
	rec := doServe(t, s, "/?format=yaml", "203.0.113.5:1234", nil)

	require.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	assert.Contains(t, body, `ip: "203.0.113.5"`)
	assert.Contains(t, body, "is_bogon: false")
}

func TestServeHTTP_FormatUnknown(t *testing.T) {
	s := newTestIPServer(false)
	rec := doServe(t, s, "/?format=xml", "203.0.113.5:1234", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "unknown format")
	assert.Contains(t, rec.Body.String(), "xml")
}

func TestServeHTTP_FormatEmptyExplicit(t *testing.T) {
	s := newTestIPServer(false)
	rec := doServe(t, s, "/?format=", "203.0.113.5:1234", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "203.0.113.5", rec.Body.String())
}

func TestServeHTTP_HeaderWithFormatJSON(t *testing.T) {
	s := newTestIPServer(false)
	headers := http.Header{
		"Cf-Connecting-Ip": []string{"198.51.100.77"},
	}
	rec := doServe(t, s, "/?format=json", "203.0.113.5:1234", headers)

	require.Equal(t, http.StatusOK, rec.Code)
	var got Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "198.51.100.77", got.IP)
}

func TestServeHTTP_DumpMode(t *testing.T) {
	s := newTestIPServer(true)
	rec := doServe(t, s, "/?format=json", "203.0.113.5:1234", http.Header{
		"User-Agent": []string{"test-agent"},
	})

	require.Equal(t, http.StatusOK, rec.Code)
	var got Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "203.0.113.5", got.IP)
}

func TestServeHTTP_MultiValueXFFTakesFirst(t *testing.T) {
	s := newTestIPServer(false)
	headers := http.Header{
		"X-Forwarded-For": []string{"  198.51.100.1 , 198.51.100.2 , 198.51.100.3 "},
	}
	rec := doServe(t, s, "/", "203.0.113.5:1234", headers)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "198.51.100.1", rec.Body.String())
}

func TestServeHTTP_OnlyAllowGETMethod(t *testing.T) {
	s := newTestIPServer(false)
	for _, method := range []string{
		http.MethodPost, http.MethodPut, http.MethodConnect,
		http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodPatch,
		http.MethodDelete,
		"BREW", "PROPFIND", "WHEN",
	} {
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.Equal(t, "GET", rec.Header().Get("Allow"))
	}
}
