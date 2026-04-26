package run

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"os/signal"
	"syscall"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/lookctx"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/lightddns"
	"github.com/duakc/mt"
	"github.com/duakc/mt/mtmap"

	goyaml "github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Arguments struct {
	Options options.Options

	Config string
	Once   bool
}

var (
	Command = &cobra.Command{
		Use:   "run",
		Short: "Run " + constpkg.Project,
		Run:   entry,
	}

	arg Arguments
)

func init() {
	Command.Flags().StringVarP(&arg.Config, "config", "c", "", "Set config file path")
	Command.Flags().BoolVar(&arg.Once, "once", false, "Trigger once and quit")
	// TODO: add a fast way to configure options rather than config file
}

func entry(cmd *cobra.Command, args []string) {
	if arg.Config == "" {
		_ = cmd.Help()
		return
	}
	if err := openConfigBind(arg.Config, &arg.Options); err != nil {
		zaplog.Fatal("read config file failed", zap.String("file", arg.Config), zap.Error(err))
		return
	}
	ctx := lookctx.NewRegistry(context.Background(), lookctx.NewDefaultRegistry())
	ddns, err := lightddns.New(ctx, arg.Options)
	if err != nil {
		zaplog.Fatal("initial instance failed", zap.Error(err))
	}
	if err := runInstance(ctx, ddns); err != nil {
		zaplog.Fatal("start failed", zap.Error(err))
	}
}

func runInstance(ctx context.Context, ddns *lightddns.LightDDNS) error {
	var cancel context.CancelFunc
	ctx, cancel = signal.NotifyContext(ctx, os.Interrupt, os.Kill,
		syscall.SIGINT, syscall.SIGHUP, syscall.SIGABRT)
	defer cancel()

	ddns.Once(ctx)
	if arg.Once {
		return nil
	}

	return ddns.Start(ctx)
}

type configTemplateContext struct {
	Env map[string]string
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
	result := make(map[string]string)

	for _, v := range os.Environ() {
		key, value, ok := mt.KeyValue(v)
		if ok {
			result[key] = value
		}
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
		result = mtmap.MergeMap(result, unmarshalBytes)
	}
	return result
}
