package gendoc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testdata1 = `@V 1 @A @B
@VersionRequired 
@DOC this 
is a doc name

@nextmeta`

func TestTokenizer(t *testing.T) {
	token := NewTokenizer(strings.NewReader(testdata1))
	for !token.NextMeta() {
		name := token.MetaName()
		switch name {
		case "@A", "@B", "@VersionRequired", "@nextmeta":
		case "@V":
			token.NextMeta()
			assert.Equal(t, "1", token.FragmentText())
		case "@DOC":
			token.NextMeta()
			assert.Equal(t, "this \nis a doc name", token.FragmentText())
		default:
			require.FailNow(t, "unexcepted meta name: "+name)
		}
	}
}
