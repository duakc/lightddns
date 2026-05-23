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
	"github.com/duakc/mt/services/filehelper"

	goyaml "github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Arguments struct {
	Options options.Options

	Config       string
	Once         bool
	OnceFastFail bool
}

var (
	Command = &cobra.Command{
		Use:   "run",
		Short: "Run " + constpkg.Project,
		Long:  "Start the DDNS updater daemon. Monitors IP changes from datasources and pushes updates to DNS providers on the configured interval.",
		Run:   entry,
	}

	commandArgument Arguments
)

func init() {
	Command.Flags().StringVarP(&commandArgument.Config, "config", "c", "", "Path to the configuration file")
	Command.Flags().BoolVar(&commandArgument.Once, "once", false, "Update all domains once and exit")
	Command.Flags().BoolVar(&commandArgument.OnceFastFail, "once-fastfail", false,
		"Exit immediately on the first update error (requires --once)")
	// TODO: add a fast way to configure options rather than config file
}

func entry(cmd *cobra.Command, args []string) {
	if commandArgument.Config == "" {
		_ = cmd.Help()
		return
	}
	ctx := globalcontext.Load()
	if err := openConfigBind(ctx, commandArgument.Config, &commandArgument.Options); err != nil {
		zaplog.Fatal("read config file failed", zap.String("file", commandArgument.Config), zap.Error(err))
	}

	ddns, err := lightddns.New(ctx, commandArgument.Options)
	if err != nil {
		zaplog.Fatal("initial instance failed", zap.Error(err))
	}
	signalCtx, cancel := gos.InterruptSignalContext(ctx)
	defer cancel()
	if commandArgument.Once {
		runInstanceOnce(signalCtx, ddns)
		return
	}

	runInstance(signalCtx, ddns)
}

func runInstanceOnce(ctx context.Context, ddns *lightddns.LightDDNS) {
	// pre-start
	if err := ddns.Start(ctx, services.StagePreStart); err != nil {
		zaplog.Fatal("start instance failed", zap.Error(err))
	}

	// once-start
	if err := ddns.StartOnce(ctx, commandArgument.OnceFastFail); err != nil {
		zaplog.Fatal("update once failed", zap.Error(err))
	}

	if err := ddns.Close(); err != nil {
		zaplog.Fatal("close instance failed", zap.Error(err))
	}
}

func runInstance(ctx context.Context, ddns *lightddns.LightDDNS) {
	if err := services.StartService(ctx, ddns); err != nil {
		zaplog.Fatal("start instance failed", zap.Error(err))
	}

	// wait
	<-ctx.Done()

	if err := services.CloseService(ddns); err != nil {
		zaplog.Warn("close instance failed", zap.Error(err))
	}
}

func openConfigBind(ctx context.Context, file string, opt *options.Options) error {
	type configTemplateContext struct {
		Env map[string]string
	}
	fileHelper := services.Lookup[filehelper.Helper](ctx)

	tempFile, err := template.ParseFiles(fileHelper.Path(file))
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
