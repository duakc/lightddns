package command

import (
	"bufio"
	"bytes"
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
	"github.com/duakc/mt/services/closeme"
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
	command := sh.New()
	command.Envs(option.Env.Values)

	cc := &Command{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),

		cmdIPv4:  xtypes.NewJoinedString(option.Cmd.IPv4.Value, " "),
		cmdIPv6:  xtypes.NewJoinedString(option.Cmd.IPv6.Value, " "),
		exitCode: option.ExitCode,

		stdin:  option.Stdin,
		stdout: option.Stdout,
		stderr: option.Stderr,
	}

	logger := adapter.CreateDatasourceLogger(zaplog.FromContext(ctx), cc)
	cc.logger = logger

	if option.Shell == "none" {
		cc.noShell = true
	} else if option.Shell != "" {
		shell, isShell := sh.ShellFromString(option.Shell)
		if !isShell {
			return nil, fmt.Errorf("unknown shell: %s", option.Shell)
		}
		command.Shell = shell
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

	closers closeme.Manager
	noShell bool

	stdinData []byte

	stdin, stdout, stderr string
}

func (c *Command) IP(ctx context.Context) ([]netip.Addr, error) {
	return adapter.MergeDualStackDatasourceIP(ctx, c)
}

func (c *Command) IPv4(ctx context.Context) ([]netip.Addr, error) {
	ip, err := runCommand(ctx, c.logger, c.cmd, c.cmdIPv4,
		c.exitCode, c.noShell, c.stdinData)
	if err != nil {
		return nil, err
	}
	return netool.FilterAddress(ip, true, false), nil
}

func (c *Command) IPv6(ctx context.Context) ([]netip.Addr, error) {
	ip, err := runCommand(ctx, c.logger, c.cmd, c.cmdIPv6,
		c.exitCode, c.noShell, c.stdinData)
	if err != nil {
		return nil, err
	}
	return netool.FilterAddress(ip, false, true), nil
}

func (c *Command) Close() error {
	return c.closers.Close()
}

func (c *Command) Start(ctx context.Context, stage services.Stage) error {
	switch stage {
	case services.StagePreStart:
		fileHelper := services.Lookup[filehelper.Helper](ctx)
		if c.stdin != "" {
			stdin, err := fileHelper.Root().ReadFile(c.stdin)
			if err != nil {
				return fmt.Errorf("open stdin: %w", err)
			}

			c.stdinData = stdin
		}
		if c.stdout != "" {
			stdout, err := fileHelper.Create(c.stdout)
			if err != nil {
				return fmt.Errorf("open stdout: %w", err)
			}
			closeme.AddClose(c.closers, stdout)
			c.cmd.Stdout = stdout
			if c.stderr == "" {
				// redirect stderr to the same stdout
				c.cmd.Stderr = stdout
			}
		}

		if c.stderr != "" {
			stderr, err := fileHelper.Create(c.stderr)
			if err != nil {
				return fmt.Errorf("open stderr: %w", err)
			}
			closeme.AddClose(c.closers, stderr)
			c.cmd.Stderr = stderr
		}
	case services.StagePostStart, services.StageStart:
		return nil
	default:
		panic("unknown stage: " + stage.String())
	}
	return nil
}

func runCommand(ctx context.Context, logger *zap.Logger,
	cmd *sh.Cmd, command xtypes.Joined[string], exitCode int, noShell bool,
	stdinData []byte,
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
	if len(stdinData) > 0 {
		runCmd.Stdin = bytes.NewReader(stdinData)
	}

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
