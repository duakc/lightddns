package main

import (
	"os"

	"github.com/duakc/lightddns/script/goscript/build"
	"github.com/duakc/lightddns/script/goscript/schema"
)

func main() {
	t := os.Args[1]
	if len(os.Args) > 2 {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	}
	switch t {
	case "build":
		build.Run()
	case "schema":
		schema.Run()
	}
}
