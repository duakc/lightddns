package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/datasources/internal"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/freebuf"
	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/filehelper"

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

func New(ctx context.Context, logger *zap.Logger, option options.CommandDatasourceOption) (adapter.Datasource, error) {
	if len(option.Cmd.Value) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	if option.Capture == "" {
		option.Capture = options.CommandOutputStdout
	}
	if option.Capture == options.CommandOutputNone {
		return nil, fmt.Errorf("set `capture` to `%s` is not allowed", options.CommandOutputNone)
	}

	stdout, stderr := io.Discard, io.Discard
	switch option.Output {
	case options.CommandOutputAll:
		stdout, stderr = os.Stdout, os.Stderr
	case options.CommandOutputStdout:
		stdout = os.Stdout
	case options.CommandOutputStderr:
		stderr = os.Stderr
	}

	rc := &runCommandContext{
		logger:   logger,
		commands: option.Cmd.Value,
		exitCode: option.ExitCode,

		stdout: stdout,
		stderr: stderr,
		env:    append(os.Environ(), option.Env...),
		matcher: internal.NewDefaultIPMatcher(
			option.Match.JQ.Cast(),
			option.Match.Regexp.Cast(),
		),
	}

	switch option.Capture {
	case options.CommandOutputAll:
		rc.captureStdout, rc.captureStderr = true, true
	case options.CommandOutputStdout:
		rc.captureStdout = true
	case options.CommandOutputStderr:
		rc.captureStderr = true
	}

	lightddnsFileHelper := services.Lookup[filehelper.Helper](ctx)

	workDir := option.WorkDir
	if workDir == "" {
		workDir = lightddnsFileHelper.Path(".")
	} else if !filepath.IsAbs(workDir) && lightddnsFileHelper != nil {
		workDir = lightddnsFileHelper.Path(workDir)
	}
	rc.workDir = workDir

	if option.Stdin != "" {
		stdinPath := option.Stdin
		if !filepath.IsAbs(stdinPath) {
			stdinPath = filepath.Join(workDir, stdinPath)
		}
		stdinBuffer, err := os.ReadFile(stdinPath)
		if err != nil {
			return nil, fmt.Errorf("read stdin file: %s: %w", option.Stdin, err)
		}
		rc.stdinBuffer = stdinBuffer
	}

	if rc.stdinBuffer == nil && option.StdinContent != "" {
		rc.stdinBuffer = []byte(option.StdinContent)
	}

	var syncMutex *sync.Mutex
	if option.Sync {
		syncMutex = &sync.Mutex{}
	}

	return &Command{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),
		logger:              logger,
		runCommandContext:   rc,
		syncMutex:           syncMutex,
	}, nil
}

type Command struct {
	adapter.AbstractManagedType

	logger *zap.Logger

	syncMutex *sync.Mutex

	runCommandContext *runCommandContext
}

func (c *Command) IP(ctx context.Context) ([]netip.Addr, error) {
	if c.syncMutex != nil {
		c.syncMutex.Lock()
		defer c.syncMutex.Unlock()
	}
	return c.runCommandContext.Handle(ctx)
}

type runCommandContext struct {
	logger *zap.Logger

	commands []string

	stdout, stderr io.Writer
	stdinBuffer    []byte

	matcher *internal.IPMatcher

	workDir string
	env     []string

	exitCode int

	captureStdout bool
	captureStderr bool
}

func (rc *runCommandContext) Handle(ctx context.Context) ([]netip.Addr, error) {
	const maxCommandOutputBufferSize = 1<<16 - 1

	buf := freebuf.NewSerialLimited(maxCommandOutputBufferSize)
	defer buf.FreeMe()

	logger := rc.logger.WithLazy(zap.Strings("commands", rc.commands))

	var (
		stdout = rc.stdout
		stderr = rc.stderr
	)

	if rc.captureStdout {
		logger.Debug("command stdout will be captured")
		stdout = io.MultiWriter(rc.stdout, buf)
	}
	if rc.captureStderr {
		logger.Debug("command stderr will be captured")
		stderr = io.MultiWriter(rc.stderr, buf)
	}

	cmd := exec.CommandContext(ctx, rc.commands[0], rc.commands[1:]...)
	cmd.Env = rc.env
	cmd.Dir = rc.workDir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if len(rc.stdinBuffer) > 0 {
		cmd.Stdin = bytes.NewReader(rc.stdinBuffer)
	}

	startErr := cmd.Start()
	if startErr != nil {
		return nil, startErr
	}

	logger.Debug("start and wait until process quit",
		zap.Int("pid", cmd.Process.Pid))

	waitErr := cmd.Wait()
	if waitErr != nil && (cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != rc.exitCode) {
		return nil, waitErr
	}

	logger.Info("quited",
		zap.String("state", cmd.ProcessState.String()))

	if cmd.ProcessState.ExitCode() != rc.exitCode {
		return nil, fmt.Errorf("unexpected exit code: %d", cmd.ProcessState.ExitCode())
	}

	return rc.matcher.Try(ctx, buf.Bytes())
}

//func runCommand(ctx context.Context, logger *zap.Logger,
//	cmd *sh.Cmd, command xtypes.Joined[string], exitCode int, noShell bool,
//	stdinData []byte,
//) ([]netip.Addr, error) {
//	const maxOutputBufferSize = 1<<16 - 1
//	logger = logger.WithLazy(zap.String("command", command.Join))
//	buf := freebuf.NewSerialLimited(maxOutputBufferSize)
//	defer buf.FreeMe()
//
//	runCmd := *cmd
//	runCmd.Stdout = io.MultiWriter(buf, runCmd.Stdout)
//	runCmd.Stderr = io.MultiWriter(buf, runCmd.Stderr)
//	if len(stdinData) > 0 {
//		runCmd.Stdin = bytes.NewReader(stdinData)
//	}
//
//	logger.Debug("run", zap.String("shell", cmd.Shell.String()))
//	var err error
//	if noShell {
//		err = runCmd.ExecCommand(ctx, command.Array[0], command.Array[1:]...).Run()
//	} else {
//		err = runCmd.RunContext(ctx, command.Join)
//	}
//	if err != nil {
//		shellErr, ok := errors.AsType[*sh.ShellError](err)
//		if !ok {
//			return nil, err
//		}
//		if code := sh.ExitCode(shellErr); code != exitCode {
//			return nil, shellErr.Unwrap()
//		}
//	}
//	logger.Debug("exit succeed", zap.Int("exit_code", exitCode))
//	if ce := logger.Check(zap.DebugLevel, "command output"); ce != nil {
//		ce.Write(zap.Int("len", buf.Len()), zap.String("output", string(buf.Bytes())))
//	}
//
//	var ans []netip.Addr
//	bufScan := bufio.NewScanner(buf)
//	for bufScan.Scan() {
//		line := bufScan.Text()
//		ipSplit := strings.Fields(line)
//		for i := 0; i < len(ipSplit); i++ {
//			ip := mt.UnquoteString(ipSplit[i])
//			addr, err := netip.ParseAddr(ip)
//			if err != nil {
//				return nil, err
//			}
//			ans = append(ans, addr)
//		}
//	}
//	return ans, nil
//}
