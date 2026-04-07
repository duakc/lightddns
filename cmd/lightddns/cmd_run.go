package main

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/duakc/lightddns"
	"github.com/duakc/lightddns/infra/common"
	"github.com/duakc/lightddns/infra/ctxservice"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	goyaml "github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type commandArgRunType struct {
	options.Options

	Config string

	Once bool
}

type configTemplateContext struct {
	Env map[string]string
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
	runCommand.Flags().BoolVar(&cmdOption.Once, "once", false, "Trigger once and quit")
	// TODO: add a fast way to configure options rather than config file

	rootCommand.AddCommand(runCommand)
}

func commandEntryRun(cmd *cobra.Command, args []string) {
	if cmdOption.Config == "" {
		_ = cmd.Help()
		return
	}
	if err := openConfigBind(cmdOption.Config, &cmdOption.Options); err != nil {
		zaplog.Fatal("read config file failed", zap.String("file", cmdOption.Config), zap.Error(err))
		return
	}

	if err := runMain(); err != nil {
		zaplog.Fatal("start failed", zap.Error(err))
	}
}

func runMain() error {
	ctx := ctxservice.NewRegistry(context.Background(), ctxservice.NewDefaultRegistry())
	ctx, stop := signal.NotifyContext(ctx,
		os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()

	ddns, err := lightddns.New(ctx, cmdOption.Options)
	if err != nil {
		return err
	}
	ddns.Once(ctx)
	if cmdOption.Once {
		return nil
	}

	return ddns.Start(ctx)
}

func openConfigBind(file string, opt *options.Options) error {
	tempFile, err := template.ParseFiles(file)
	if err != nil {
		return err
	}
	var tempContext configTemplateContext
	tempContext.Env = fullEnv()
	configBuffer := bytes.NewBuffer(nil)
	if err := tempFile.Execute(configBuffer, tempContext); err != nil {
		return err
	}

	decoder := goyaml.NewDecoder(configBuffer, goyaml.DisallowUnknownField())
	return decoder.Decode(opt)
}

func fullEnv() map[string]string {
	var result = make(map[string]string)

	for _, v := range os.Environ() {
		vv := strings.SplitN(v, "=", 2)
		result[vv[0]] = vv[1]
	}
	if _, err := os.Stat(".env"); err == nil {
		envFile, err := os.ReadFile(".env")
		if err != nil {
			zaplog.Warn(".env file found but read failed", zap.Error(err))
			return result
		}
		unmarshalBytes, err := godotenv.UnmarshalBytes(envFile)
		if err != nil {
			zaplog.Warn(".env file parse failed", zap.Error(err))
			return result
		}
		result = common.MergeMap(result, unmarshalBytes)
	}
	return result
}
