package check

import (
	"github.com/duakc/lightddns/cmd/lightddns/internal/globalcontext"
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/lightddns"
	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/filehelper"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Arguments struct {
	Config string
}

var (
	commandArguments = &Arguments{}
	Command          = &cobra.Command{
		Use:   "check",
		Short: "Check configuration validity",
		Long:  "Parse the configuration file, create datasources and providers, then exit. Use this to validate config before deploying.",
		Run:   entry,
	}
)

func init() {
	Command.Flags().StringVarP(&commandArguments.Config, "config", "c", "", "Path to configuration file")
}

func entry(cmd *cobra.Command, args []string) {
	ctx := globalcontext.Load()
	fileHelper := services.Lookup[filehelper.Helper](ctx)
	configFile, err := fileHelper.Open(commandArguments.Config)
	if err != nil {
		zaplog.Fatal("open config failed", zap.Error(err))
	}
	defer configFile.Close()
	var option options.Options

	if err = badyaml.NewDecoder(configFile).Decode(&option); err != nil {
		zaplog.Fatal("decode config failed", zap.Error(err))
	}

	if _, err = lightddns.New(ctx, option); err != nil {
		zaplog.Fatal("create new ddns instance failed", zap.Error(err))
	}
}
