package filehelper

import (
	"os"
	"path/filepath"

	"github.com/duakc/mt"
)

var osRoot = mt.Must(os.OpenRoot("."))

func WorkingDir(wd string) {
	err := createDir(wd)
	if err != nil {
		panic(err)
	}
	root, err := os.OpenRoot(wd)
	if err != nil {
		panic(err)
	}
	osRoot = root
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if err := createDir(name); err != nil {
		return nil, err
	}
	return osRoot.OpenFile(name, flag, perm)
}

func Create(name string) (*os.File, error) {
	if err := createDir(name); err != nil {
		return nil, err
	}
	return osRoot.Create(name)
}

func Open(name string) (*os.File, error) {
	return osRoot.Open(name)
}

func createDir(name string) error {
	dir := filepath.Dir(name)
	return osRoot.MkdirAll(dir, 0o777)
}

func Root() *os.Root {
	return osRoot
}
