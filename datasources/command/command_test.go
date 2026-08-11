package command

import (
	"context"
	"fmt"
	"io"
	"net/netip"
	"os"
	"testing"

	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/filehelper"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRejectsInvalidCommandOptions(t *testing.T) {
	t.Run("empty command", func(t *testing.T) {
		_, err := New(context.Background(), zap.NewNop(), options.CommandDatasourceOption{})
		require.ErrorContains(t, err, "empty command")
	})

	t.Run("capture none", func(t *testing.T) {
		_, err := New(context.Background(), zap.NewNop(), commandOption("stdout", func(o *options.CommandDatasourceOption) {
			o.Capture = options.CommandOutputNone
		}))
		require.ErrorContains(t, err, "set `capture` to `none` is not allowed")
	})
}

func TestCommandIPCapturesStdoutByDefault(t *testing.T) {
	ctx, _ := commandTestContext(t)
	ds, err := New(ctx, zap.NewNop(), commandOption("stdout"))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("192.0.2.1", "2001:db8::1"), got)
}

func TestCommandIPCaptureSelectsStderr(t *testing.T) {
	ctx, _ := commandTestContext(t)
	ds, err := New(ctx, zap.NewNop(), commandOption("stderr", func(o *options.CommandDatasourceOption) {
		o.Capture = options.CommandOutputStderr
	}))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("198.51.100.7"), got)
}

func TestCommandIPUsesDefaultWorkDirFromContext(t *testing.T) {
	ctx, root := commandTestContext(t)
	require.NoError(t, os.WriteFile(root+"/ip.txt", []byte("203.0.113.10\n"), 0o600))

	ds, err := New(ctx, zap.NewNop(), commandOption("workdir"))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.10"), got)
}

func TestCommandIPUsesConfiguredWorkDir(t *testing.T) {
	ctx, _ := commandTestContext(t)
	workDir := t.TempDir()
	require.NoError(t, os.WriteFile(workDir+"/ip.txt", []byte("203.0.113.11\n"), 0o600))

	ds, err := New(ctx, zap.NewNop(), commandOption("workdir", func(o *options.CommandDatasourceOption) {
		o.WorkDir = workDir
	}))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.11"), got)
}

func TestCommandIPResolvesRelativeWorkDirFromContext(t *testing.T) {
	ctx, root := commandTestContext(t)
	workDir := root + "/nested"
	require.NoError(t, os.Mkdir(workDir, 0o700))
	require.NoError(t, os.WriteFile(workDir+"/ip.txt", []byte("203.0.113.16\n"), 0o600))

	ds, err := New(ctx, zap.NewNop(), commandOption("workdir", func(o *options.CommandDatasourceOption) {
		o.WorkDir = "nested"
	}))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.16"), got)
}

func TestCommandIPUsesStdinFileBeforeStdinContent(t *testing.T) {
	ctx, _ := commandTestContext(t)
	workDir := t.TempDir()
	require.NoError(t, os.WriteFile(workDir+"/stdin.txt", []byte("203.0.113.12\n"), 0o600))

	ds, err := New(ctx, zap.NewNop(), commandOption("stdin", func(o *options.CommandDatasourceOption) {
		o.WorkDir = workDir
		o.Stdin = "stdin.txt"
		o.StdinContent = "198.51.100.12\n"
	}))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.12"), got)
}

func TestCommandIPUsesStdinContent(t *testing.T) {
	ctx, _ := commandTestContext(t)
	ds, err := New(ctx, zap.NewNop(), commandOption("stdin", func(o *options.CommandDatasourceOption) {
		o.StdinContent = "203.0.113.13\n"
	}))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.13"), got)
}

func TestCommandIPUsesConfiguredEnvironment(t *testing.T) {
	ctx, _ := commandTestContext(t)
	ds, err := New(ctx, zap.NewNop(), commandOption("env", func(o *options.CommandDatasourceOption) {
		o.Env = append(o.Env, "LIGHTDDNS_TEST_IP=203.0.113.14")
	}))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.14"), got)
}

func TestCommandIPAcceptsExpectedNonZeroExitCode(t *testing.T) {
	ctx, _ := commandTestContext(t)
	ds, err := New(ctx, zap.NewNop(), commandOption("exit", func(o *options.CommandDatasourceOption) {
		o.ExitCode = 7
	}))
	require.NoError(t, err)

	got, err := ds.IP(ctx)
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.15"), got)
}

func TestCommandIPRejectsUnexpectedExitCode(t *testing.T) {
	ctx, _ := commandTestContext(t)
	ds, err := New(ctx, zap.NewNop(), commandOption("exit"))
	require.NoError(t, err)

	_, err = ds.IP(ctx)
	require.Error(t, err)
}

func commandOption(mode string, apply ...func(*options.CommandDatasourceOption)) options.CommandDatasourceOption {
	opt := options.CommandDatasourceOption{
		Cmd: badyaml.Listable[string]{
			Value: []string{os.Args[0], "-test.run=^TestCommandHelperProcess$", "--", mode},
		},
		Env: []string{"LIGHTDDNS_COMMAND_HELPER=1"},
	}
	for _, fn := range apply {
		fn(&opt)
	}
	return opt
}

func commandTestContext(t *testing.T) (context.Context, string) {
	t.Helper()

	root := t.TempDir()
	helper, err := filehelper.New(root)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, helper.Close())
	})

	ctx := services.NewRegistry(context.Background(), services.NewDefaultRegistry())
	return services.Store[filehelper.Helper](ctx, helper), root
}

func parseAddrs(values ...string) []netip.Addr {
	addrs := make([]netip.Addr, 0, len(values))
	for _, value := range values {
		addrs = append(addrs, netip.MustParseAddr(value))
	}
	return addrs
}

func TestCommandHelperProcess(t *testing.T) {
	if os.Getenv("LIGHTDDNS_COMMAND_HELPER") != "1" {
		return
	}

	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "missing command helper mode")
		os.Exit(2)
	}

	switch args[1] {
	case "stdout":
		fmt.Fprint(os.Stdout, "192.0.2.1\n2001:db8::1\n")
	case "stderr":
		fmt.Fprint(os.Stderr, "198.51.100.7\n")
	case "workdir":
		data, err := os.ReadFile("ip.txt")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Fprint(os.Stdout, string(data))
	case "stdin":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Fprint(os.Stdout, string(data))
	case "env":
		fmt.Fprint(os.Stdout, os.Getenv("LIGHTDDNS_TEST_IP"))
	case "exit":
		fmt.Fprint(os.Stdout, "203.0.113.15\n")
		os.Exit(7)
	default:
		fmt.Fprintf(os.Stderr, "unknown command helper mode %q\n", args[1])
		os.Exit(2)
	}
	os.Exit(0)
}
