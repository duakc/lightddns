package httpx

import (
	"mime"
	"net/http"
	"strings"
)

func ToStandardMethod(method string) (string, bool) {
	method = strings.ToUpper(method)
	switch method {
	case "", http.MethodGet, http.MethodPost, http.MethodPut, http.MethodConnect,
		http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodPatch,
		http.MethodDelete,
		"BREW", "PROPFIND", "WHEN": // RFC2324
		return method, true
	default:
		return "", false
	}
}

func IsJsonContentType(contentType string) bool {
	mediatype, _, err := mime.ParseMediaType(contentType)
	return err == nil && (mediatype == "application/json" || mediatype == "text/json")
}

func ExtendHeaders(source, extended http.Header) {
	for k, v := range extended {
		for _, vv := range v {
			source.Add(k, vv)
		}
	}
}

func ExtendHeadersOverride(source, extended http.Header) {
	for k, v := range extended {
		for _, vv := range v {
			source.Set(k, vv)
		}
	}
}
