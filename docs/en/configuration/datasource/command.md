# Command

Runs shell commands to discover IP addresses. Each command's stdout is parsed line-by-line for valid IP addresses.

```yaml
# required
type: command
name: data-cmd

cmd:
  ipv4: "curl -s https://api.ipify.org"
  ipv6: "curl -s https://api6.ipify.org"

# optional
shell: ""
exitCode: 0
env:
  PATH: "/usr/bin:/bin:/usr/sbin:/sbin"
stdin: ""
stdout: ""
stderr: ""
```

??? note "Behavior"
    IPv4 and IPv6 commands run independently. When a command is empty, that IP version is skipped. Command output is read line-by-line; each whitespace-separated token on a line is parsed as an IP address. Only addresses matching the requested version are returned.

---

## `cmd`

The commands to run for IPv4 and IPv6. When a command is empty, that IP version is skipped.

**Short form** (`string`) — same command for both IPv4 and IPv6:

```yaml
cmd: "curl -s https://api.ipify.org"
```

**Object form** — separate commands for IPv4 and IPv6:

```yaml
cmd:
  ipv4: "curl -s https://api.ipify.org"
  ipv6: "curl -s https://api6.ipify.org"
```

**Array form** — pass the program and arguments as a list. Required when using `shell: "none"`:

```yaml
cmd:
  ipv4: ["curl", "-s", "https://api.ipify.org"]
  ipv6: ["curl", "-s", "https://api6.ipify.org"]
```

---

## `shell`

The shell interpreter used to run the commands. When empty, a default is selected automatically.

Set to `"none"` to run the command directly without any shell. This avoids [shell injection](https://en.wikipedia.org/wiki/Code_injection#Shell_injection) and word-splitting issues, but **requires the array form of `cmd`** — otherwise the command string is treated as the executable name, which will fail.

!!! warning "`shell: none`"
    When `shell` is set to `none`, you must use the **array form** of `cmd`. A plain string like `"curl -s https://api.ipify.org"` will be treated as a single executable name, not a command with arguments.

```yaml
# Without a shell — array form is required
shell: none
cmd:
  ipv4: ["curl", "-s", "https://api.ipify.org"]
```

??? note "Supported shells by platform"
    Windows:
    - powershell, cmd

    Linux:
    - bash, zsh, fish, dash, sh, ash, mksh, csh, tcsh, rksh, ksh

    macOS:
    - zsh, bash, sh, ksh, csh, tcsh, fish, dash, ash, mksh

    FreeBSD / OpenBSD / NetBSD / Dragonfly:
    - sh, csh, tcsh, bash, zsh, fish, mksh, dash, ash, rksh, ksh

```yaml
shell: bash
```

---

## `exitCode`

The expected exit code for a successful run. All other exit codes are treated as errors.

```yaml
exitCode: 0
```

---

## `env`

Environment variables for the command execution.

```yaml
env:
  PATH: "/usr/bin:/bin:/usr/sbin:/sbin"
```

---

## `stdin`

Path to a file whose contents are piped to the command's stdin. If the command needs continuous input, use a streaming file such as a domain socket.

!!! warning Stdin  
    The functionality of this configuration option currently needs improvement, 
    and the exact wording may change in the future. For cases requiring password input, 
    such as `sudo`, I highly recommend using files like `/etc/sudoers` to configure it.

```yaml
stdin: /path/to/input.txt
```

## `stdout`

Path to a file where the command's stdout is written. IP addresses are still parsed from stdout as normal. If `stderr` is not set, stderr is redirected to the same file.

```yaml
stdout: /var/log/ddns-ip.log
```

## `stderr`

Path to a file where the command's stderr is written.

```yaml
stderr: /var/log/ddns-ip-error.log
```
