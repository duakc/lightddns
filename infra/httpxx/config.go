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
)

type AcceptCodeFunc func(code int) bool

type ReqConfig struct {
	Method  string
	BaseURL *urlpkg.URL

	Query        urlpkg.Values
	ExtendPath   []string
	ExtendHeader http.Header

	// inputs order:
	Body any

	// AcceptStatus reports whether a response status code is acceptable.
	// If nil, the default is "any status code < 400".
	AcceptStatus AcceptCodeFunc
}

// Accepts reports whether code is acceptable according to AcceptStatus.
// When AcceptStatus is nil, the default "< 400" rule is used. This lets
// downstream consumers and tests stay agnostic about whether the field
// was explicitly set.
func (rc ReqConfig) Accepts(code int) bool {
	if rc.AcceptStatus == nil {
		return code < 400
	}
	return rc.AcceptStatus(code)
}

func StatusAcceptEqual(codec int) AcceptCodeFunc {
	return func(code int) bool {
		return code == codec
	}
}

func StatusAcceptGreater(codec int) AcceptCodeFunc {
	return func(code int) bool {
		return code > codec
	}
}

func StatusAcceptLess(codec int) AcceptCodeFunc {
	return func(code int) bool {
		return code < codec
	}
}

func NewReqConfig(method string, baseURL *urlpkg.URL) ReqConfig {
	return ReqConfig{
		Method:       method,
		BaseURL:      baseURL,
		Query:        make(urlpkg.Values),
		ExtendHeader: make(http.Header),
	}
}

// ToRequestContext builds an *http.Request from rc. It does not mutate rc:
// ExtendHeader and Query are read-only; any header that needs to be added
// (e.g. an inferred Content-Type) is written to a clone, never the original.
func (rc ReqConfig) ToRequestContext(ctx context.Context) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if rc.BaseURL == nil {
		return nil, fmt.Errorf("BaseURL is required")
	}

	u := *rc.BaseURL
	if len(rc.ExtendPath) != 0 {
		u.Path = path.Join(append([]string{u.Path}, rc.ExtendPath...)...)
	}
	if len(rc.Query) > 0 {
		q := u.Query()
		for k, v := range rc.Query {
			for _, vv := range v {
				q.Add(k, vv)
			}
		}
		u.RawQuery = q.Encode()
	}

	header := rc.ExtendHeader.Clone()
	if header == nil {
		header = make(http.Header)
	}

	var body io.Reader
	if rc.Body != nil && slices.Contains([]string{
		http.MethodPost, http.MethodPut, http.MethodPatch,
	}, rc.Method) {
		defaultContentType := ""
		switch x := rc.Body.(type) {
		case json.Marshaler:
			defaultContentType = "application/json"
			data, err := x.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("marshal JSON: %w", err)
			}
			body = bytes.NewBuffer(data)
		case encoding.BinaryMarshaler:
			defaultContentType = "application/octet-stream"
			data, err := x.MarshalBinary()
			if err != nil {
				return nil, fmt.Errorf("marshal binary: %w", err)
			}
			body = bytes.NewBuffer(data)
		case encoding.TextMarshaler:
			defaultContentType = "text/plain"
			text, err := x.MarshalText()
			if err != nil {
				return nil, fmt.Errorf("marshal text: %w", err)
			}
			body = bytes.NewBuffer(text)
		case io.Reader:
			if header.Get("Content-Type") == "" {
				return nil, fmt.Errorf("undetermined Content-Type")
			}
			body = x
		}
		if defaultContentType != "" && header.Get("Content-Type") == "" {
			header.Set("Content-Type", defaultContentType)
		}
	}

	req, err := http.NewRequestWithContext(ctx, rc.Method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header = header
	return req, nil
}
