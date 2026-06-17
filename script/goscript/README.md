# goscript — design spec

`goscript` is this project's build/release tooling, written in **Go instead of
shell**. If you expected shell scripts under `script/`, this note explains why
it's Go and the conventions to keep when changing it.

## Why Go, not shell

- **One source of truth, in real code.** The cross-compilation matrix (every
  GOOS/GOARCH and its packaging support) lives in the `target` package and is
  consumed by both building and packaging. No duplicated lists, no manifest to
  drift out of sync.
- **Portable and reproducible.** No bash/`sed`/`awk` portability traps; it runs
  the same on a developer's macOS laptop and on Linux CI.
- **Reuses the project's own code and types**, and is testable like any Go
  package — far easier to maintain than a growing pile of shell.

## Conventions

Follow these when adding or changing a command.

1. **Orchestrate in Go; run the real tool.** A command prepares the
   *preconditions* and *data*, then shells out (`os/exec`) to the tool that does
   the real work — `build` runs `go build`, `deb` runs `dpkg-deb`. Do **not**
   reimplement a tool's job in Go (no hand-written `ar`/`tar`/`.deb` format); the
   tool owns its format.

2. **`target` is the single source of truth.** It holds the build matrix, the
   binary-naming rule, and — declared right in the table — whether each row ships
   as a `.deb`/`.rpm` (`Deb`/`RPM` flags). Add an architecture in one place and
   it flows everywhere. The flags are pre-filled even for commented-out,
   not-yet-enabled rows, so enabling a rare OS/arch later needs no packaging
   research. Arch *names* differ per packaging system, so they come from one
   small translation function, never repeated per row.

3. **Put each concern where it belongs.** Data only meaningful to `go build`
   (the build environment) lives in `build`, not on `Target`. What is intrinsic
   to a target (its name, packaging support) lives on `Target`.

4. **Build first, then package.** `build.Binary` is the reusable unit that
   compiles one target; packaging consumes a binary, so it calls `build.Binary`
   before invoking the packaging tool.

5. **Keep abstractions lean.** Prefer a few small, obvious helpers and the
   standard library over a deep utility layer.

## Structure

`run.go` is a thin dispatcher: it switches on the first argument and calls the
matching package's `Run(ctx context.Context)`.

- **Every top-level package under `goscript/` is a command** (`build`, `deb`,
  `gendoc`, `genschema`, ...) with a `Run(ctx)` that parses its own flags.
- **Shared, non-command libraries live under `pkg/`** (`run` for executing
  external commands, `gitver` for version detection, `target` for the build
  matrix, `jsonschema`).
- **External commands always go through `pkg/run`**, so each invocation is
  echoed (with `-verbose`) for debugging and there's one place to extend.

To add a command: create `script/goscript/<name>` with a `Run(ctx)` and add a
`case` to `run.go`. The package *payload* commands operate on (deb skeleton,
systemd units, the generated `release/schema.json`) stays under `release/` as
plain, tracked files.