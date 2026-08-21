package main

import (
	"context"
	"flag"
	"os"

	"github.com/duakc/lightddns/infra/gos"
	"github.com/duakc/lightddns/script/goscript/build"
	"github.com/duakc/lightddns/script/goscript/gendoc"
	"github.com/duakc/lightddns/script/goscript/genschema"
	"github.com/duakc/lightddns/script/goscript/nfpmpkg"
	"github.com/duakc/lightddns/script/goscript/pkg/buildinfo"
	"github.com/duakc/lightddns/script/goscript/pkg/gobuild"
)

// Thin dispatcher:
//
// `go run run.go [flags...] <command> [flags...]`
func main() {
	//args := parseGlobalFlags(os.Args[1:])
	//if len(args) == 0 {
	//	return
	//}

	// re-shape os.Args so each command's flag.Parse sees only its own flags.
	command := os.Args[1]
	os.Args = os.Args[1:]

	buildinfo.RegisterFlags(flag.CommandLine)
	gobuild.RegisterBuildFlags(flag.CommandLine)

	ctx, cancel := gos.InterruptSignalContext(context.Background())
	defer cancel()
	switch command {
	case "build":
		build.Run(ctx)
	case "gendoc":
		gendoc.Run(ctx)
	case "genschema":
		genschema.Run(ctx)
	case "nfpm":
		nfpmpkg.Run(ctx)
	default:
		panic("unknown command: " + command)
	}
}

//func parseGlobalFlags(args []string) []string {
//	i := 0
//	for i < len(args) && strings.HasPrefix(args[i], "-") {
//		name, val, hasEq := strings.Cut(args[i], "=")
//		value := func() string { // value for the "--flag VALUE" form
//			if hasEq {
//				return val
//			}
//			if i+1 < len(args) {
//				i++
//				return args[i]
//			}
//			return ""
//		}
//		switch name {
//		default:
//			return args[i:] // unknown flag: leave it for the sub-command
//		}
//		i++
//	}
//	return args[i:]
//}
