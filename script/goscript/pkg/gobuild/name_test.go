package gobuild

import (
	"testing"

	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/stretchr/testify/assert"
)

func TestBinaryName(t *testing.T) {
	tgt := target.Target{GOOS: "linux", GOARCH: "amd64"}
	params := Params{BinaryName: "lightddns", Qualified: true}

	assert.Equal(t, "lightddns-linux-amd64", binaryName(tgt, params, "(unknown)"))
	assert.Equal(t, "lightddns-linux-amd64-v1.2.3", binaryName(tgt, params, "v1.2.3"))
	assert.Equal(t, "lightddns-v1.2.3", binaryName(tgt, Params{BinaryName: "lightddns"}, "v1.2.3"))
	assert.Equal(t, "lightddns-windows-amd64-v1.2.3.exe",
		binaryName(target.Target{GOOS: "windows", GOARCH: "amd64"}, params, "v1.2.3"))
}
