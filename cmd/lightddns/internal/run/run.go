package run

import (
	"bytes"
	"context"
	"html/template"
	"os"

	"github.com/duakc/lightddns/cmd/lightddns/internal/globalcontext"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/gos"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/lightddns"
	"github.com/duakc/mt"
	"github.com/duakc/mt/mtmap"
	"github.com/duakc/mt/services"

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
	}
	ctx := globalcontext.Load()

	ddns, err := lightddns.New(ctx, arg.Options)
	if err != nil {
		zaplog.Fatal("initial instance failed", zap.Error(err))
	}

	runInstance(ctx, ddns)
}

func runInstance(ctx context.Context, ddns *lightddns.LightDDNS) {
	ctx, cancel := gos.InterruptSignalContext(ctx)
	defer cancel()

	err := services.StartService(ctx, ddns)
	if err != nil {
		zaplog.Fatal("start service failed", zap.Error(err))
	}
	// wait
	<-ctx.Done()
	err = services.CloseService(ddns)
	if err != nil {
		zaplog.Warn("close failed", zap.Error(err))
	}
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
