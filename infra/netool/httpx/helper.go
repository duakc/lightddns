package httpx

import (
	"bytes"
	"fmt"
	"io"
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
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && (mediaType == "application/json" ||
		mediaType == "text/json" || strings.HasSuffix(mediaType, "+json"))
}

func ExtendHeaders(source, extended http.Header) {
	for k, v := range extended {
		for _, vv := range v {
			source.Add(k, vv)
		}
	}
}

func ExtendHeadersOverride(source, extended http.Header) {
	for key, values := range extended {
		source[key] = append([]string(nil), values...)
	}
}

func ReadAndReplayBody(request *http.Request) ([]byte, error) {
	if request == nil {
		return nil, fmt.Errorf("read and replay body: nil request")
	}

	var body []byte
	if request.Body != nil {
		original := request.Body
		var err error
		body, err = io.ReadAll(original)
		closeErr := original.Close()
		if err != nil {
			return nil, fmt.Errorf("read and replay body: %w", err)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close request body: %w", closeErr)
		}
	}

	request.Body = io.NopCloser(bytes.NewReader(body))
	request.ContentLength = int64(len(body))
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	return body, nil
}
