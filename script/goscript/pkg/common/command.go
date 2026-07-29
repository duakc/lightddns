package common

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var Verbose = os.Getenv("GOSCRIPT_VERBOSE") == "1"

type Cmd struct {
	Name string
	Args []string
	Env  []string // extra variables, appended to os.Environ()
	Dir  string
}

func CommandStream(ctx context.Context, c Cmd) error {
	cmd := c.prepare(ctx)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

func CommandCapture(ctx context.Context, c Cmd) (string, error) {
	out, err := c.prepare(ctx).Output()
	return strings.TrimSpace(string(out)), err
}

func (c Cmd) prepare(ctx context.Context) *exec.Cmd {
	if Verbose {
		// Command echo stays uncoloured - it's a faithful transcript.
		_, _ = fmt.Fprintf(os.Stderr, "$ %s\n", c)
	}
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	if len(c.Env) > 0 {
		cmd.Env = append(os.Environ(), c.Env...)
	}
	return cmd
}

func (c Cmd) String() string {
	parts := append([]string{}, c.Env...)
	parts = append(parts, c.Name)
	for _, a := range c.Args {
		if strings.ContainsRune(a, ' ') {
			a = "'" + a + "'"
		}
		parts = append(parts, a)
	}
	return strings.Join(parts, " ")
}
