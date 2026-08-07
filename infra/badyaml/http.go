package badyaml

import (
	"fmt"
	"net/http"
	"strings"
)

type HTTPMethod string

func (m *HTTPMethod) UnmarshalYAML(data []byte) error {
	s, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	method := strings.ToUpper(s)
	switch method {
	case "", http.MethodGet, http.MethodPost, http.MethodPut, http.MethodConnect,
		http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodPatch,
		http.MethodDelete,
		"BREW", "PROPFIND", "WHEN": // RFC2324
		*m = HTTPMethod(method)
	default:
		return fmt.Errorf("unknown http method: %s", method)
	}
	return nil
}

type HTTPHeader struct {
	Header http.Header
}

func (h *HTTPHeader) UnmarshalYAML(data []byte) error {
	m := make(map[string]any)
	if err := Unmarshal(data, &m); err != nil {
		return err
	}
	h.Header = make(http.Header)
	for k, v := range m {
		switch val := v.(type) {
		case []any:
			for _, item := range val {
				if s, ok := item.(string); ok {
					h.Header.Add(k, s)
				} else {
					h.Header.Add(k, fmt.Sprint(item))
				}
			}
		default:
			h.Header.Add(k, fmt.Sprint(v))
		}
	}
	return nil
}
