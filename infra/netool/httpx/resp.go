package httpx

import (
	"fmt"
	"net/http"
	"net/netip"
	"strings"
)

type RespPolicy struct {
	AcceptCode func(code int) bool
}

func (rp RespPolicy) AcceptResponse(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}

	if !rp.acceptCodes(resp.StatusCode) {
		return &BadStatusCodeError{Got: resp.StatusCode}
	}

	return nil
}

func (rp RespPolicy) acceptCodes(code int) bool {
	if rp.AcceptCode == nil {
		return code < 400
	}
	return rp.AcceptCode(code)
}

func ExtractIPFromRequest(req *http.Request) ([]netip.Addr, error) {
	extract := func(s string) []netip.Addr {
		if s == "" {
			return nil
		}
		var ips []netip.Addr
		for _, part := range strings.Split(s, ",") {
			ipStr := strings.TrimSpace(part)
			if ipStr == "" {
				continue
			}
			if addr, err := netip.ParseAddr(ipStr); err == nil {
				ips = append(ips, addr)
			}
		}
		return ips
	}

	for _, header := range []string{
		"Cf-Connecting-IP",
		"True-Client-IP",
		"X-Real-IP",
		"X-Forwarded-For",
	} {
		if ips := extract(req.Header.Get(header)); len(ips) > 0 {
			return ips, nil
		}
	}

	addrPort, err := netip.ParseAddrPort(req.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid remote address: %s: %w", req.RemoteAddr, err)
	}
	return []netip.Addr{addrPort.Addr()}, nil
}

type StatusCodeResponseWriter struct {
	http.ResponseWriter

	statusCode int
}

func (w *StatusCodeResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *StatusCodeResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
