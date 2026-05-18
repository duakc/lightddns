# Command

通过执行 Shell 命令获取 IP 地址。命令的标准输出将逐行解析，提取合法的 IP 地址。

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

??? note "行为说明"
    IPv4 和 IPv6 命令独立执行。命令为空时跳过对应的 IP 版本。命令输出逐行读取，每行中的空白分隔标记均被解析为 IP 地址。仅返回与请求版本匹配的地址。

---

## `cmd`

分别为 IPv4 和 IPv6 指定执行命令。命令为空时，跳过对应的 IP 版本。

**简写形式**（`string`）— IPv4 和 IPv6 共用同一个命令：

```yaml
cmd: "curl -s https://api.ipify.org"
```

**对象形式** — 单独为 IPv4 和 IPv6 指定命令：

```yaml
cmd:
  ipv4: "curl -s https://api.ipify.org"
  ipv6: "curl -s https://api6.ipify.org"
```

**数组形式** — 将程序和参数分开传入，使用 `shell: "none"` 时必须用此形式：

```yaml
cmd:
  ipv4: ["curl", "-s", "https://api.ipify.org"]
  ipv6: ["curl", "-s", "https://api6.ipify.org"]
```

---

## `shell`

用于执行命令的 shell 解释器。为空时会自动选择。

设为 `"none"` 时直接执行命令，不经过任何 shell。这可以避免[命令注入](https://en.wikipedia.org/wiki/Code_injection#Shell_injection)和字符串拆分的隐患，但**必须搭配数组形式**的 `cmd` 使用——否则整个命令字符串会被当作可执行文件名，导致运行失败。

!!! warning "`shell: none`"
    使用 `shell: none` 时，`cmd` **必须**使用数组形式。`"curl -s https://api.ipify.org"` 这样的字符串会被视为一个整体文件名，而非"命令 + 参数"。

```yaml
# 不使用 shell — 必须用数组形式
shell: none
cmd:
  ipv4: ["curl", "-s", "https://api.ipify.org"]
```

??? note "各平台支持的 Shell"
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

期望的命令退出码。其他退出码将被视为执行失败。

```yaml
exitCode: 0
```

---

## `env`

命令执行环境的变量。

```yaml
env:
  PATH: "/usr/bin:/bin:/usr/sbin:/sbin"
```

---

## `stdin`

将指定文件的内容作为命令的标准输入。若命令需要持续输入，可使用流式传输文件，如 domain socket。

!!! warning Stdin
    这个配置项现在功能是需要改进的，具体行文也许会在未来变化。
    如果为了 `sudo` 这类的需要输入密码的，我更推荐使用 `/etc/sudoers` 这类文件去配置。

```yaml
stdin: /path/to/input.txt
```

## `stdout`

将命令的标准输出写入指定文件。IP 地址仍然会从标准输出中正常解析。若未设置 `stderr`，标准错误也会重定向到同一文件。

```yaml
stdout: /var/log/ddns-ip.log
```

## `stderr`

将命令的标准错误写入指定文件。

```yaml
stderr: /var/log/ddns-ip-error.log
```
