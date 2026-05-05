package main

import (
	"context"
	"os"

	"github.com/duakc/lightddns/infra/gos"
	"github.com/duakc/lightddns/script/goscript/build"
	"github.com/duakc/lightddns/script/goscript/gendoc"
)

func main() {
	t := os.Args[1]
	if len(os.Args) > 2 {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	}
	ctx, cancel := gos.InterruptSignalContext(context.Background())
	defer cancel()
	switch t {
	case "build":
		build.Run(ctx)
	case "gendoc":
		gendoc.Run(ctx)
	}
}
