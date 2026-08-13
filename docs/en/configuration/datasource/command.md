This document was translated from Chinese by AI.

# Command

Obtains IP addresses by running an external command and extracting addresses from its output.

```yaml
# required
type: command
name: data-cmd
cmd: ["curl", "-s", "https://api.ipify.org"]

# optional
exitCode: 0
env:
  - PATH=/usr/bin:/bin:/usr/sbin:/sbin
output: none
capture: stdout
stdin: ""
stdinContent: ""
sync: false
workDir: ""
match:
  regex: ""
```

??? note "Behavior"
    The command is executed directly using the configured argument list. Lightddns does not add an implicit shell. Array form is recommended for commands with arguments. To use shell syntax such as pipes, redirection, environment variable expansion, or `&&`, invoke a shell explicitly.

---

## `cmd`

The command to run.

!!! note
    A command without arguments does not need to use array form.

```yaml
cmd: ["curl", "-s", "https://api.ipify.org"]
```

To use shell syntax, invoke a shell explicitly:

```yaml
cmd: ["sh", "-c", "curl -s https://api.ipify.org | tr -d '\\n'"]
```

## `exitCode`

The expected command exit code. Any other exit code is treated as an error.

```yaml
exitCode: 0
```

## `env`

Environment variables appended to the command execution environment. Environment variables inherited from the Lightddns process are retained, and the configured `KEY=VALUE` entries are appended.

```yaml
env:
  - PATH=/usr/bin:/bin:/usr/sbin:/sbin
  - TOKEN={{ .Env.MY_TOKEN }}
```

## `output`

Controls which output streams are forwarded to Lightddns's own stdout or stderr so that command execution can be observed. It does not control which output stream is parsed.

| Value           | Behavior                                |
|-----------------|-----------------------------------------|
| `none` or empty | Do not forward command output.          |
| `stdout`        | Forward stdout.                         |
| `stderr`        | Forward stderr.                         |
| `all`           | Forward both stdout and stderr.         |

```yaml
output: stderr
```

## `capture`

Controls which output streams are captured and parsed.

| Value    | Behavior                                |
|----------|-----------------------------------------|
| empty    | Equivalent to `stdout`.                 |
| `stdout` | Parse stdout.                           |
| `stderr` | Parse stderr.                           |
| `all`    | Parse both stdout and stderr.           |

```yaml
capture: stdout
```

## `stdin`

Uses the contents of the specified file as the command's standard input.

Relative paths are resolved from the effective working directory of this datasource:

1. If `workDir` is configured, use `workDir`.
2. Otherwise, use the global working directory set by `lightddns -D/--workdir`.
3. If `-D` is not passed explicitly, its default is `.`, the current working directory of the process.

Absolute paths are read as-is. If both `stdin` and `stdinContent` are set, `stdin` takes precedence.

```yaml
stdin: input.txt
```

## `stdinContent`

Inline standard input content. This field is ignored when `stdin` is set.

```yaml
stdinContent: |
  query payload
```

## `sync`

When set to `true`, concurrent IP queries for the same datasource are executed serially. This can be enabled when the command reads or writes shared local state.

??? note
    This option mainly prevents mixed stdout output during concurrent execution, but it makes the command datasource fully single-threaded.

```yaml
sync: true
```

## `workDir`

The working directory of the command process and the base directory for relative `stdin` paths.

When empty, the global working directory set by `lightddns -D/--workdir` is used. A relative value is resolved from the global working directory. An absolute path is used as-is.

```yaml
workDir: scripts
```

For example, when running `lightddns -D /etc/lightddns run -c config.yaml`, the configuration above runs the command under `/etc/lightddns/scripts`.

## `match`

IP extraction rules. See [MatchOption](../shared/match.md).

```yaml
match:
  jq: ".ip"
  regex: "IP:\\s+(\\S+)"
```
