package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strings"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	datasourcepkg "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/lookctx"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"
	"github.com/duakc/mt/freebuf"
	"github.com/duakc/mt/sh"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const DatasourceType = constpkg.DatasourceTypeCommand

func init() {
	adapter.Register(
		adapter.DatasourceRegister,
		DatasourceType,
		New,
	)
}

func New(ctx context.Context, option options.CommandDatasourceOption) (adapter.Datasource, error) {
	logger := datasourcepkg.NewLogger(lookctx.LookupPtr[zap.Logger](ctx), option.AbstractDatasourceOption)
	command := sh.New()
	if option.Shell != "" {
		shell, isShell := sh.ShellFromString(option.Shell)
		if !isShell {
			return nil, fmt.Errorf("unknown shell: %s", option.Shell)
		}
		command = sh.NewShell(shell)
	}

	command.Deattach()
	command.Stdin = os.Stdin
	command.Envs(option.Env.Values)

	return &Command{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),

		logger: logger,
		cmd:    command,

		cmdIPv4:  option.CmdV4,
		cmdIPv6:  option.CmdV6,
		exitCode: option.ExitCode,
	}, nil
}

type Command struct {
	adapter.AbstractManagedType

	logger *zap.Logger

	cmd      *sh.Cmd
	cmdIPv4  string
	cmdIPv6  string
	exitCode int
}

func (c *Command) IP(ctx context.Context) ([]netip.Addr, error) {
	return adapter.MergeDualStackDatasourceIP(ctx, c)
}

func (c *Command) IPv4(ctx context.Context) ([]netip.Addr, error) {
	ip, err := runCommand(ctx, c.logger, c.cmd, c.cmdIPv4, c.exitCode)
	if err != nil {
		return nil, err
	}
	return netool.FilterAddress(ip, true, false), nil
}

func (c *Command) IPv6(ctx context.Context) ([]netip.Addr, error) {
	ip, err := runCommand(ctx, c.logger, c.cmd, c.cmdIPv6, c.exitCode)
	if err != nil {
		return nil, err
	}
	return netool.FilterAddress(ip, false, true), nil
}

func runCommand(ctx context.Context, logger *zap.Logger,
	cmd *sh.Cmd, command string, exitCode int,
) ([]netip.Addr, error) {
	const maxOutputBufferSize = 1 << 16

	if command == "" {
		return []netip.Addr{}, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	logger = logger.WithLazy(zap.String("command", command))
	buf := freebuf.New(maxOutputBufferSize)
	defer buf.FreeMe()

	runCmd := *cmd
	runCmd.Stdout = buf
	runCmd.Stderr = buf

	logger.Debug("run", zap.String("shell", cmd.Shell.String()))
	err := runCmd.RunContext(ctx, command)
	if err != nil {
		shellErr, ok := errors.AsType[*sh.ShellError](err)
		if !ok {
			return nil, err
		}
		if code := sh.ExitCode(shellErr); code != exitCode {
			return nil, shellErr.Unwrap()
		}
	}
	logger.Debug("exit succeed", zap.Int("exit_code", exitCode))
	if ce := logger.Check(zapcore.TraceLevel, "command output"); ce != nil {
		ce.Write(zap.Int("len", buf.Len()), zap.String("output", string(buf.Bytes())))
	}

	var ans []netip.Addr
	bufScan := bufio.NewScanner(buf)
	for bufScan.Scan() {
		line := bufScan.Text()
		ipSplit := strings.Fields(line)
		for i := 0; i < len(ipSplit); i++ {
			ip := mt.UnquoteString(ipSplit[i])
			addr, err := netip.ParseAddr(ip)
			if err != nil {
				return nil, err
			}
			ans = append(ans, addr)
		}
	}
	return ans, nil
}
