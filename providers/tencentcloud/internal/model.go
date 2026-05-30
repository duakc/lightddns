package internal

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/duakc/mt/common/validator"
)

type Common struct {
	// https://cloud.tencent.com/document/api/1427/56188
	Action string
	// generate by call Headers
	// Timestamp string
	Version string

	// optional
	Token     string
	Language  string
	Region    string
	Timestamp int64
}

func (c *Common) Headers() (http.Header, error) {
	if err := errors.Join(
		validator.NonEmpty(c.Action, "Action"),
		validator.NonEmpty(c.Version, "Version"),
		validator.GreaterThan(c.Timestamp, 0, "Timestamp"),
	); err != nil {
		return nil, err
	}

	h := make(http.Header)
	h.Set("X-TC-Action", c.Action)
	h.Set("X-TC-Version", c.Version)
	h.Set("X-TC-Timestamp", fmt.Sprintf("%d", c.Timestamp))

	if len(c.Token) > 0 {
		h.Set("X-TC-Token", c.Token)
	}
	if len(c.Language) > 0 {
		h.Set("X-TC-Language", c.Language)
	}
	if len(c.Region) > 0 {
		h.Set("X-TC-Region", c.Region)
	}

	return h, nil
}
