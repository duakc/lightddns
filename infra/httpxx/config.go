package httpxx

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"mime"
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
		var err error
		body, err = buildBody(rc.Body, header)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, rc.Method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header = header
	return req, nil
}

// buildBody decides how to feed rc.Body to net/http.
//
// Routing (first match wins):
//
//  1. Content-Length is preset by the caller — the caller committed to a
//     byte count, so Body must be an io.Reader.
//
//  2. Body is an io.Reader — hand off to net/http directly. The caller
//     owns framing and Content-Type.
//
//  3. Content-Type is preset by the caller — marshal according to that
//     Content-Type. Prefer the matching Marshaler interface if the body
//     implements it; otherwise fall back to a generic encoder where one
//     exists (json.Marshal for application/json — handles map[string]any,
//     untagged structs, etc).
//
//  4. Content-Type is unset — infer it from the body's Marshaler interface.
//     If no Marshaler matches, return an error because the Content-Type
//     cannot be determined.
//
// The marshal paths produce a *bytes.Buffer so net/http can populate
// Content-Length from Buffer.Len.
func buildBody(body any, header http.Header) (io.Reader, error) {
	if header.Get("Content-Length") != "" {
		r, ok := body.(io.Reader)
		if !ok {
			return nil, fmt.Errorf(
				"Content-Length set in ExtendHeader requires io.Reader body, got %T", body)
		}
		return r, nil
	}

	if r, ok := body.(io.Reader); ok {
		return r, nil
	}

	if ct := header.Get("Content-Type"); ct != "" {
		return marshalForContentType(body, ct)
	}

	return marshalAndInferContentType(body, header)
}

func marshalForContentType(body any, ct string) (io.Reader, error) {
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, fmt.Errorf("parse Content-Type %q: %w", ct, err)
	}

	var data []byte
	switch mediaType {
	case "application/json", "text/json":
		if jm, ok := body.(json.Marshaler); ok {
			data, err = jm.MarshalJSON()
		} else {
			data, err = json.Marshal(body)
		}
		if err != nil {
			return nil, fmt.Errorf("marshal JSON: %w", err)
		}
	case "application/octet-stream":
		bm, ok := body.(encoding.BinaryMarshaler)
		if !ok {
			return nil, fmt.Errorf(
				"body %T does not implement encoding.BinaryMarshaler for Content-Type %s",
				body, ct)
		}
		data, err = bm.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("marshal binary: %w", err)
		}
	case "text/plain":
		tm, ok := body.(encoding.TextMarshaler)
		if !ok {
			return nil, fmt.Errorf(
				"body %T does not implement encoding.TextMarshaler for Content-Type %s",
				body, ct)
		}
		data, err = tm.MarshalText()
		if err != nil {
			return nil, fmt.Errorf("marshal text: %w", err)
		}
	default:
		return nil, fmt.Errorf(
			"cannot marshal body %T for Content-Type %s; provide an io.Reader", body, ct)
	}
	return bytes.NewBuffer(data), nil
}

func marshalAndInferContentType(body any, header http.Header) (io.Reader, error) {
	switch x := body.(type) {
	case json.Marshaler:
		data, err := x.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("marshal JSON: %w", err)
		}
		header.Set("Content-Type", "application/json")
		return bytes.NewBuffer(data), nil
	case encoding.BinaryMarshaler:
		data, err := x.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("marshal binary: %w", err)
		}
		header.Set("Content-Type", "application/octet-stream")
		return bytes.NewBuffer(data), nil
	case encoding.TextMarshaler:
		data, err := x.MarshalText()
		if err != nil {
			return nil, fmt.Errorf("marshal text: %w", err)
		}
		header.Set("Content-Type", "text/plain")
		return bytes.NewBuffer(data), nil
	default:
		return nil, fmt.Errorf(
			"cannot determine Content-Type for body %T; "+
				"set Content-Type in ExtendHeader, implement a Marshaler interface, "+
				"or pass an io.Reader", body)
	}
}
