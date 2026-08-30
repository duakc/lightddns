package run

import (
	"github.com/duakc/lightddns/cmd/lightddns/internal/common"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/gos"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/mt"

	"github.com/duakc/lightddns"
	"github.com/duakc/mt/services"

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
		Long:  "Start the DDNS updater daemon. Monitors IP changes from datasources and pushes updates to DNS providers on the configured interval.",
		Run:   entry,
	}

	commandArgument Arguments
)

func init() {
	Command.Flags().StringVarP(&commandArgument.Config, "config", "c", "", "Path to the configuration file")
	Command.Flags().BoolVar(&commandArgument.Once, "once", false, "Update all domains once and exit")
	// TODO: add a fast way to configure options rather than config file
}

func entry(cmd *cobra.Command, args []string) {
	if commandArgument.Config == "" {
		_ = cmd.Help()
		return
	}
	ctx := common.Context()

	opt, err := common.LoadConfig(ctx, commandArgument.Config)
	if err != nil {
		zaplog.Fatal("read config file failed", zap.String("file", commandArgument.Config), zap.Error(err))
	}
	commandArgument.Options = opt

	ddns, err := lightddns.New(ctx, commandArgument.Options)
	if err != nil {
		zaplog.Fatal("initial instance failed", zap.Error(err))
	}

	defer func() {
		closeErr := ddns.Close()
		if closeErr != nil {
			zaplog.Fatal("close failed", zap.Error(closeErr))
		}
	}()

	// seal
	services.RegistryFromContext(ctx).Seal()

	signalCtx, cancel := gos.InterruptSignalContext(ctx)
	defer cancel()
	if commandArgument.Once {
		ddns.SetOnce()
	} else {
		defer mt.WaitContext(signalCtx)
	}

	if err := services.StartService(signalCtx, ddns); err != nil {
		zaplog.Fatal("start instance failed", zap.Error(err))
	}
}
