package httpxx

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
