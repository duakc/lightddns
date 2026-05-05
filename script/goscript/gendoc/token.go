package gendoc

import (
	"bytes"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

const MetaPrefix = "@"

var metaPrefixRune, _ = utf8.DecodeRune([]byte(MetaPrefix))

type Tokenizer struct {
	reader io.Reader
	ioErr  error

	buf []byte
	r   int
	b   int

	// indicate the buf end
	e int

	// indicate that the read progressed but an error occurred
	// dirty >= r
	// err == nil => dirty == -1
	dirty int
}

func NewTokenizerString(s string) *Tokenizer {
	return NewTokenizer(strings.NewReader(s))
}

func NewTokenizer(r io.Reader) *Tokenizer {
	t := &Tokenizer{reader: r}
	t.Reset()
	return t
}

func (t *Tokenizer) Err() error {
	e := t.ioErr
	t.ioErr = nil
	return e
}

func (t *Tokenizer) Reset() {
	t.buf = nil
	t.r, t.b = 0, 0
	t.dirty = -1
}

func (t *Tokenizer) NextMeta() bool {
	return t.readUntil(metaPrefixRune, nil)
}

func (t *Tokenizer) MetaName() string {
	if !t.readUntil(0, unicode.IsSpace) {
		frag := t.Fragment()
		if id := bytes.IndexRune(t.Fragment(), metaPrefixRune); id >= 0 {
			return string(bytes.TrimSpace(frag[id:]))
		}
		return ""
	}

	return string(t.Fragment())
}

func (t *Tokenizer) Fragment() []byte {
	return t.buf[t.b:t.r]
}

func (t *Tokenizer) FragmentText() string {
	frag := t.Fragment()

	return string(bytes.TrimSpace(frag))
}

func (t *Tokenizer) readUntil(until rune, fn func(r rune) bool) bool {
	t.b = t.r
	for t.fill() {
		if t.r == t.e {
			return false
		}
		if t.ioErr != nil && t.ioErr != io.EOF {
			t.dirty = t.r
			return false
		}
		if fn != nil {
			if id := bytes.IndexFunc(t.buf[t.r:t.e], fn); id >= 0 {
				t.r += id
				return true
			}
		} else {
			if id := bytes.IndexRune(t.buf[t.r:t.e], until); id >= 0 {
				t.r += id
				return true
			}
		}
		t.r = t.e
	}
	return false
}

func (t *Tokenizer) fill() bool {
	// data before t.b are discarded
	if t.b > 0 {
		offset := t.b
		t.b = 0
		// discard useless data, try to alloc more free space
		copy(t.buf[0:t.e], t.buf[offset:t.e])
		t.r -= offset
		t.e -= offset
	}

	if t.ioErr == nil && t.e == len(t.buf) {
		// no more free space in buf, but need more free space, how:
		// resize
		buf := make([]byte, t.nextSize())
		copy(buf[:t.e], t.buf[:t.e])
		t.buf = buf
	}

	if t.ioErr == nil {
		nn, err := t.reader.Read(t.buf[t.e:])
		t.e = t.e + nn
		t.ioErr = err
	}

	if t.ioErr == io.EOF {
		return t.e > t.r
	}
	return t.ioErr == nil
}

func (t *Tokenizer) nextSize() int {
	const minSize = 4096    // 4k
	const maxSize = 1 << 20 // 1mb
	curSize := len(t.buf)
	if curSize < minSize {
		return minSize
	}

	plus := curSize
	if curSize >= maxSize {
		plus = maxSize
	}
	return curSize + plus
}
