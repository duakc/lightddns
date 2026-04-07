package badyaml

import (
	"bytes"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strings"

	"github.com/duakc/lightddns/infra/common"
	"github.com/duakc/lightddns/infra/netxx"
	goyaml "github.com/goccy/go-yaml"
)

type HTTPMethod string

func (m *HTTPMethod) UnmarshalYAML(data []byte) error {
	method := strings.ToUpper(common.UnquoteString(string(data)))
	switch method {
	case "", http.MethodGet, http.MethodPost, http.MethodPut, http.MethodConnect,
		http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodPatch,
		http.MethodDelete, "BREW", "PROPFIND", "WHEN": // RFC2324
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
	decoder := goyaml.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(m)
	if err != nil {
		return err
	}
	h.Header = make(http.Header)
	for k, v := range m {
		switch val := v.(type) {
		case string:
			val = common.UnquoteString(val)
			h.Header.Add(k, val)
		case []any:
			for _, item := range val {
				if s, ok := item.(string); ok {
					h.Header.Add(k, common.UnquoteString(s))
				} else {
					h.Header.Add(k, common.UnquoteString(fmt.Sprintf(s)))
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
	url := common.UnquoteString(string(data))
	parse, err := urlpkg.Parse(url)
	if err != nil {
		return err
	}
	m.URL = parse
	m.Raw = url
	return nil
}

type DomainName string

func (d *DomainName) UnmarshalYAML(data []byte) error {
	s := common.UnquoteString(string(data))
	if !netxx.IsDomainName(s) {
		return fmt.Errorf("invalid domain name: %s", s)
	}
	*d = DomainName(s)
	return nil
}
