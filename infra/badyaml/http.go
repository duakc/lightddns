package badyaml

import (
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strings"

	"github.com/duakc/lightddns/infra/netool"
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
		case string:
			h.Header.Add(k, val)
		case []any:
			for _, item := range val {
				if s, ok := item.(string); ok {
					h.Header.Add(k, s)
				} else {
					h.Header.Add(k, fmt.Sprint(item))
				}
			}
		}
	}
	return nil
}

type URL struct {
	URL *urlpkg.URL
	Raw string
}

func (m *URL) UnmarshalYAML(data []byte) error {
	s, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	parse, err := urlpkg.Parse(s)
	if err != nil {
		return err
	}
	m.URL = parse
	m.Raw = s
	return nil
}

type DomainName string

func (d *DomainName) UnmarshalYAML(data []byte) error {
	s, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	if !netool.IsDomainName(s) {
		return fmt.Errorf("invalid domain name: %s", s)
	}
	*d = DomainName(s)
	return nil
}
