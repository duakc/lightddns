package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"unicode"
	"unsafe"

	"github.com/itchyny/gojq"
)

type IPMatcher struct {
	JQ     *gojq.Query
	Regexp *regexp.Regexp

	Parser func(c string) (netip.Addr, error)
}

func NewDefaultIPMatcher(jq *gojq.Query, regexp *regexp.Regexp) *IPMatcher {
	return &IPMatcher{
		JQ:     jq,
		Regexp: regexp,
		Parser: netip.ParseAddr,
	}
}

func (m *IPMatcher) Try(ctx context.Context, content []byte) ([]netip.Addr, error) {
	var (
		ans []netip.Addr

		jqErr, reErr, plainErr error
	)

	if m.JQ != nil {
		ans, jqErr = m.JSON(ctx, content)
		if jqErr == nil && len(ans) > 0 {
			return ans, nil
		}
	}
	if m.Regexp != nil {
		ans, reErr = m.Re(content)
		if reErr == nil && len(ans) > 0 {
			return ans, nil
		}
	}
	ans, plainErr = m.Plain(content)
	if len(ans) > 0 {
		return ans, nil
	}
	return nil, errors.Join(jqErr, reErr, plainErr)
}

func (m *IPMatcher) Plain(content []byte) ([]netip.Addr, error) {
	var joinedErr error

	parts := findPart(content)
	ans := make([]netip.Addr, 0, len(parts))
	for _, part := range parts {
		parsed, err := m.Parser(unsafe.String(unsafe.SliceData(part), len(part)))
		if err != nil {
			joinedErr = errors.Join(joinedErr, err)
			continue
		}
		ans = append(ans, parsed)
	}
	if len(ans) > 0 {
		return ans, nil
	}
	return nil, joinedErr
}

func (m *IPMatcher) JSON(ctx context.Context, content []byte) ([]netip.Addr, error) {
	var jsonObject any
	if err := json.Unmarshal(content, &jsonObject); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}
	var ans []netip.Addr
	iter := m.JQ.RunWithContext(ctx, jsonObject)

	for val, ok := iter.Next(); ok; val, ok = iter.Next() {
		switch x := val.(type) {
		case error:
			if haltErr, ok := errors.AsType[*gojq.HaltError](x); ok && haltErr.Value() == nil {
				break
			}
			return nil, fmt.Errorf("jq execution error: %w", x)
		case string:
			addr, err := m.Parser(x)
			if err != nil {
				return nil, err
			}
			ans = append(ans, addr)
		}
	}

	return ans, nil
}

func (m *IPMatcher) Re(content []byte) ([]netip.Addr, error) {
	var ans []netip.Addr

	for _, x := range m.Regexp.FindAllSubmatch(content, -1) {
		if len(x) <= 1 {
			continue
		}

		xx := x[1]
		addr, err := m.Parser(unsafe.String(unsafe.SliceData(xx), len(xx)))
		if err != nil {
			return nil, err
		}
		ans = append(ans, addr)
	}
	return ans, nil
}

func findPart(content []byte) [][]byte {
	content = bytes.TrimSpace(content)

	var ans [][]byte

	for i := 0; i < len(content); {
		j := bytes.IndexFunc(content[i:], func(r rune) bool {
			return !unicode.IsSpace(r)
		})

		if j == -1 {
			break
		}

		i += j
		start := i

		k := bytes.IndexFunc(content[i:], unicode.IsSpace)
		if k == -1 {
			ans = append(ans, content[start:])
			break
		}
		i += k
		ans = append(ans, content[start:i])
	}

	return ans
}
