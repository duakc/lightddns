package packing

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"unsafe"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"
)

type SubSetVec = []string

func subSetName(v string) string {
	return fmt.Sprintf("__%s__", v)
}

type SubSet map[string]string

func (s SubSet) Process(vec SubSetVec, data []byte) ([]byte, error) {
	for _, k := range vec {
		fullSubSetName := subSetName(k)

		var (
			subsetValue   string
			subsetExisted bool
		)

		if subsetValue, subsetExisted = s[k]; !subsetExisted {
			return nil, fmt.Errorf("subset %s does not exist", k)
		}

		data = bytes.ReplaceAll(data,
			unsafe.Slice(unsafe.StringData(fullSubSetName), len(fullSubSetName)),
			unsafe.Slice(unsafe.StringData(subsetValue), len(subsetValue)),
		)
	}
	return data, nil
}

// File describes one file to pack: read From the FS root, optionally
// placeholder-substituted, written To the stage tree as Mode. Declare it as a
// plain literal; the source bytes are read lazily on first Process.
type File struct {
	// FS is the source root; From is the path within it.
	FS   filehelper.Helper
	From string

	// To is where the file lands in the stage tree (relative to its root).
	To string

	Mode os.FileMode

	SubSetVec SubSetVec

	// Gzip gzips the (substituted) bytes before writing - e.g. man pages.
	Gzip bool

	once   sync.Once
	source []byte
}

// read lazily loads (and caches) the source bytes, so a File can be declared at
// package scope without touching the filesystem until it is actually packed.
func (file *File) read() []byte {
	file.once.Do(func() {
		file.source = file.FS.MustReadFile(mt.Must(filepath.Localize(file.From)))
	})
	return file.source
}

func (file *File) Process(stageFileHelper filehelper.Helper, subSet SubSet) error {
	replacedBuffer, err := subSet.Process(file.SubSetVec, slices.Clone(file.read()))
	if err != nil {
		return err
	}

	if file.Gzip {
		if replacedBuffer, err = gzipBytes(replacedBuffer); err != nil {
			return err
		}
	}

	if err := stageFileHelper.MkdirAll(filepath.Dir(file.To), 0o755); err != nil {
		return err
	}

	return stageFileHelper.WriteFile(file.To, replacedBuffer, file.Mode.Perm())
}

func gzipBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
