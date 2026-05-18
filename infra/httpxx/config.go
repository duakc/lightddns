package httpxx

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"path"
	"slices"
	"strings"
)

type ReqConfig struct {
	Method  string
	BaseURL string

	Query        urlpkg.Values
	ExtendPath   []string
	ExtendHeader http.Header

	// inputs order:
	Body any
}

func NewReqConfig(method string, baseURL string) ReqConfig {
	return ReqConfig{
		Method:       method,
		BaseURL:      baseURL,
		Query:        make(urlpkg.Values),
		ExtendHeader: make(http.Header),
	}
}

func (rc ReqConfig) ToRequestContext(ctx context.Context) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rc.BaseURL = strings.TrimSuffix(rc.BaseURL, "/") // remove trailing slash
	url := rc.BaseURL + "/" + path.Join(rc.ExtendPath...)
	if q := rc.Query.Encode(); q != "" {
		url += "?" + q
	}

	var body io.Reader
	contentType := rc.ExtendHeader.Get("Content-Type")
	if rc.Body != nil && slices.Contains([]string{
		http.MethodPost, http.MethodPut, http.MethodPatch,
	}, rc.Method) {
		switch x := rc.Body.(type) {
		case json.Marshaler:
			if contentType == "" {
				contentType = "application/json"
			}
			data, err := x.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("marshal JSON: %w", err)
			}
			body = bytes.NewBuffer(data)
		case encoding.BinaryMarshaler:
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			data, err := x.MarshalBinary()
			if err != nil {
				return nil, fmt.Errorf("marshal binary: %w", err)
			}
			body = bytes.NewBuffer(data)
		case encoding.TextMarshaler:
			if contentType == "" {
				contentType = "text/plain"
			}
			text, err := x.MarshalText()
			if err != nil {
				return nil, fmt.Errorf("marshal text: %w", err)
			}
			body = bytes.NewBuffer(text)
		case io.Reader:
			if contentType == "" {
				return nil, fmt.Errorf("undermined Content-Type")
			}
			body = x
		}
	}

	req, err := http.NewRequestWithContext(ctx, rc.Method, url, body)
	if err != nil {
		return nil, err
	}
	for h, v := range rc.ExtendHeader {
		for i := 0; i < len(v); i++ {
			req.Header.Add(h, v[i])
		}
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	return req, nil
}
