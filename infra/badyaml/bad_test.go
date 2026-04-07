package badyaml

import (
	"bytes"
	"testing"

	goyaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListable(t *testing.T) {

	type Case[T any] struct {
		Input    string
		Excepted T
	}

	type schema struct {
		Payload Listable[string] `yaml:"payload"`
	}

	cases := []Case[schema]{
		{
			Input: `
payload:
  - e1
  - "e2"
`,
			Excepted: schema{
				Payload: Listable[string]{[]string{"e1", "e2"}}},
		},
		{
			Input:    "payload: ['e1','e2','e3']",
			Excepted: schema{Payload: Listable[string]{[]string{"e1", "e2", "e3"}}},
		},
		{
			Input:    "payload: e1",
			Excepted: schema{Payload: Listable[string]{[]string{"e1"}}},
		},
	}
	for i := 0; i < len(cases); i++ {
		var vv schema
		cc := cases[i]
		err := goyaml.Unmarshal([]byte(cc.Input), &vv)
		assert.NoError(t, err)
		assert.Equal(t, cc.Excepted, vv)
	}
}

func TestBadHTTPMethod_UnmarshalYAML(t *testing.T) {
	data := `
payload:
  - "GET"
  - 'POST'
  - 'WHEN'
  - HEAD
`
	type schema struct {
		Payload []HTTPMethod `yaml:"payload"`
	}
	de := goyaml.NewDecoder(bytes.NewReader([]byte(data)))
	require.NoError(t, de.Decode(&schema{}))
}
