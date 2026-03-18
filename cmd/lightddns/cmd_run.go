package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/duakc/lightddns"
	"github.com/duakc/lightddns/options"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type commandArgRunType struct {
	options.Options

	Config string
}

var (
	runCommand = &cobra.Command{
		Use: "run",
		Run: commandEntryRun,
	}

	cmdOption commandArgRunType
)

func init() {
	runCommand.Flags().StringVarP(&cmdOption.Config, "config", "c", "", "Set config file path")
	// TODO: add a fast way to configure options rather than config file
}

func commandEntryRun(cmd *cobra.Command, args []string) {
	if cmdOption.Config == "" {
		_ = cmd.Help()
		return
	}
	if err := openConfigBind(cmdOption.Config, &cmdOption.Options); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "err: %s", err.Error())
		return
	}

	if err := runMain(); err != nil {
		panic(err)
	}
}

func runMain() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	ddns, err := lightddns.New(ctx, cmdOption.Options)
	_ = ddns
	_ = err

	return nil
}

func openConfigBind(file string, opt *options.Options) error {
	open, err := os.Open(file)
	if err != nil {
		return err
	}
	defer open.Close()
	decoder := yaml.NewDecoder(open)
	decoder.KnownFields(true)
	return decoder.Decode(opt)
}
