package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"strings"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"
	"github.com/duakc/mt/freebuf"
	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/filehelper"
	"github.com/duakc/mt/sh"
	"github.com/duakc/mt/xtypes"

	"go.uber.org/zap"
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
	logger := option.AbstractDatasourceOption.CreateLogger(zaplog.FromContext(ctx))
	fileHelper := services.Lookup[filehelper.Helper](ctx)

	command := sh.New()
	command.Envs(option.Env.Values)

	cc := &Command{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),

		logger: logger,

		cmdIPv4:  xtypes.NewJoinedString(option.Cmd.IPv4.Value, " "),
		cmdIPv6:  xtypes.NewJoinedString(option.Cmd.IPv6.Value, " "),
		exitCode: option.ExitCode,
	}

	if option.Shell == "none" {
		cc.noShell = true
	} else if option.Shell != "" {
		shell, isShell := sh.ShellFromString(option.Shell)
		if !isShell {
			return nil, fmt.Errorf("unknown shell: %s", option.Shell)
		}
		command.Shell = shell
	}

	if option.Stdin != "" {
		// TODO: make input stdin become more stable
		stdin, err := fileHelper.Open(option.Stdin)
		if err != nil {
			return nil, fmt.Errorf("open stdin: %w", err)
		}
		cc.closers = append(cc.closers, stdin)
		command.Stdin = stdin
	}

	if option.Stdout != "" {
		stdout, err := fileHelper.Create(option.Stdout)
		if err != nil {
			return nil, fmt.Errorf("open stdout: %w", err)
		}
		cc.closers = append(cc.closers, stdout)
		command.Stdout = stdout
		if option.Stderr == "" {
			// redirect stderr to the same stdout
			command.Stderr = stdout
		}
	}

	if option.Stderr != "" {
		stderr, err := fileHelper.Create(option.Stderr)
		if err != nil {
			return nil, fmt.Errorf("open stderr: %w", err)
		}
		cc.closers = append(cc.closers, stderr)
		command.Stderr = stderr
	}

	cc.cmd = command
	return cc, nil
}

type Command struct {
	adapter.AbstractManagedType

	logger *zap.Logger

	cmd     *sh.Cmd
	cmdIPv4 xtypes.Joined[string]
	cmdIPv6 xtypes.Joined[string]

	exitCode int

	closers []io.Closer
	noShell bool
}

func (c *Command) IP(ctx context.Context) ([]netip.Addr, error) {
	return adapter.MergeDualStackDatasourceIP(ctx, c)
}

func (c *Command) IPv4(ctx context.Context) ([]netip.Addr, error) {
	ip, err := runCommand(ctx, c.logger, c.cmd, c.cmdIPv4, c.exitCode, c.noShell)
	if err != nil {
		return nil, err
	}
	return netool.FilterAddress(ip, true, false), nil
}

func (c *Command) IPv6(ctx context.Context) ([]netip.Addr, error) {
	ip, err := runCommand(ctx, c.logger, c.cmd, c.cmdIPv6, c.exitCode, c.noShell)
	if err != nil {
		return nil, err
	}
	return netool.FilterAddress(ip, false, true), nil
}

func (c *Command) Close() error {
	var err error
	for _, closer := range c.closers {
		err = errors.Join(err, closer.Close())
	}
	return err
}

func runCommand(ctx context.Context, logger *zap.Logger,
	cmd *sh.Cmd, command xtypes.Joined[string], exitCode int, noShell bool,
) ([]netip.Addr, error) {
	const maxOutputBufferSize = 1 << 16

	if command.Join == "" {
		return []netip.Addr{}, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	logger = logger.WithLazy(zap.String("command", command.Join))
	buf := freebuf.New(maxOutputBufferSize)
	defer buf.FreeMe()

	runCmd := *cmd
	runCmd.Stdout = io.MultiWriter(buf, runCmd.Stdout)
	runCmd.Stderr = io.MultiWriter(buf, runCmd.Stderr)

	logger.Debug("run", zap.String("shell", cmd.Shell.String()))
	var err error
	if noShell {
		err = runCmd.ExecCommand(ctx, command.Array[0], command.Array[1:]...).Run()
	} else {
		err = runCmd.RunContext(ctx, command.Join)
	}
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
	if ce := logger.Check(zap.DebugLevel, "command output"); ce != nil {
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
