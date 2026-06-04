package internal

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Common struct {
	SecretSecurityToken string

	Nonoc string

	Time time.Time
}

func (c Common) Headers() http.Header {
	if c.Time.IsZero() {
		c.Time = time.Now()
	}
	if c.Nonoc == "" {
		c.Nonoc = uuid.NewString()
	}

	header := make(http.Header)
	header.Set(HeaderDate, c.Time.UTC().Format(time.RFC3339))
	header.Set(HeaderSignatureNonce, c.Nonoc)
	if c.SecretSecurityToken != "" {
		header.Set(HeaderSecurityToken, c.SecretSecurityToken)
	}
	return header
}
