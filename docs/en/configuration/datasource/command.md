# Command

Runs an external command and extracts IP addresses from its output.

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
    The command is executed directly with the configured argument list. Lightddns does not add an implicit shell. Use array form for commands with arguments. If you need shell features such as pipes, redirects, environment expansion, or `&&`, call the shell explicitly.

    `capture` decides which stream is parsed for IP addresses. `output` only decides which stream is also forwarded to the Lightddns process stdout/stderr.

---

## `cmd`

The command to execute.

**String form** is accepted for single executable names:

```yaml
cmd: my-ip-helper
```

**Array form** is the recommended form for commands with arguments:

```yaml
cmd: ["curl", "-s", "https://api.ipify.org"]
```

To use shell syntax, call a shell explicitly:

```yaml
cmd: ["sh", "-c", "curl -s https://api.ipify.org | tr -d '\\n'"]
```

---

## `exitCode`

The expected exit code for a successful run. Any other exit code is treated as an error.

```yaml
exitCode: 0
```

---

## `env`

Additional environment variables for the command. They are appended to the inherited process environment and use `KEY=VALUE` entries.

```yaml
env:
  - PATH=/usr/bin:/bin:/usr/sbin:/sbin
  - TOKEN={{ .Env.MY_TOKEN }}
```

---

## `output`

Controls which streams are forwarded to Lightddns stdout/stderr for visibility. It does not decide which stream is parsed.

| Value | Behavior |
|---|---|
| `none` or empty | Do not forward command output. |
| `stdout` | Forward stdout. |
| `stderr` | Forward stderr. |
| `all` | Forward both stdout and stderr. |

```yaml
output: stderr
```

---

## `capture`

Controls which streams are captured and parsed for IP addresses.

| Value | Behavior |
|---|---|
| empty | Same as `stdout`. |
| `stdout` | Parse stdout. |
| `stderr` | Parse stderr. |
| `all` | Parse both stdout and stderr. |

```yaml
capture: stdout
```

---

## `stdin`

Path to a file whose contents are piped to the command's stdin.

Relative paths are resolved against this datasource's effective working directory:

1. `workDir`, when set.
2. Otherwise the global working directory from `lightddns -D/--workdir`.
3. Otherwise the process working directory, because `-D` defaults to `.`.

Absolute paths are read as-is. When both `stdin` and `stdinContent` are set, `stdin` takes priority.

```yaml
stdin: input.txt
```

---

## `stdinContent`

Inline content to pipe to the command's stdin. Ignored when `stdin` is set.

```yaml
stdinContent: |
  query payload
```

---

## `sync`

When `true`, concurrent IP lookups for this datasource are serialized. Use this when the command reads or writes shared local state.

```yaml
sync: true
```

---

## `workDir`

Working directory for the command process and the base directory for relative `stdin` paths.

When empty, this uses the global working directory configured with `lightddns -D/--workdir`. When set to a relative path, it is resolved under that global working directory. Absolute paths are used as-is.

```yaml
workDir: scripts
```

With `lightddns -D /etc/lightddns run -c config.yaml`, the example above runs the command in `/etc/lightddns/scripts`.

---

## `match`

Optional extraction rules. See [MatchOption](../shared/match.md).

```yaml
match:
  jq: ".ip"
  regex: "IP:\\s+(\\S+)"
```
