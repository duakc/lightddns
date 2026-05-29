package internal

import (
	"net/http"

	"github.com/duakc/mt/common/validator"
)

type Common struct {
	// https://cloud.tencent.com/document/api/1427/56188
	Action string
	// generate by call Headers
	// Timestamp string
	Version string

	Authorization string

	// optional
	Token    string
	Language string
	Region   string
}

func (c *Common) Headers() (http.Header, error) {
	v := validator.NewGenericValidator(validator.DisallowEmpty[string]())
	v.Valid(c.Action, "Action")
	v.Valid(c.Version, "Version")
	v.Valid(c.Authorization, "Authorization")
	if err := v.Err(); err != nil {
		return nil, err
	}

	h := make(http.Header)
	h.Set("X-TC-Action", c.Action)
	h.Set("X-TC-Version", c.Version)
	h.Set("X-TC-Authorization", c.Authorization)

	panic("uncompleted")
}
