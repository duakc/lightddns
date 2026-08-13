package gitver

import (
	"context"
	"testing"
)

func TestBranchUsesEnvironmentOverride(t *testing.T) {
	t.Setenv(EnvBranch, "release")

	if got := branch(context.Background()); got != "release" {
		t.Fatalf("branch() = %q, want %q", got, "release")
	}
}
