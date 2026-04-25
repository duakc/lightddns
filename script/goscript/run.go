package main

import (
	"os"

	"goscript/build"
)

func main() {
	t := os.Args[1]
	if len(os.Args) > 2 {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	}
	switch t {
	case "build":
		build.Run()
	}
}
