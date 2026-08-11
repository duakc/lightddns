# Command

通过执行外部命令获取 IP 地址，并从命令输出中提取地址。

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

??? note "行为说明"
    命令会按配置的参数列表直接执行，Lightddns 不会额外包一层隐式 shell。带参数的命令建议使用数组形式。若需要管道、重定向、环境变量展开、`&&` 等 shell 语法，请显式调用 shell。

    `capture` 决定解析哪个输出流里的 IP 地址。`output` 只决定哪些输出流会同时转发到 Lightddns 进程自己的 stdout/stderr。

---

## `cmd`

要执行的命令。

**字符串形式**适合只有可执行文件名、没有参数的命令：

```yaml
cmd: my-ip-helper
```

**数组形式**适合带参数的命令，也是推荐形式：

```yaml
cmd: ["curl", "-s", "https://api.ipify.org"]
```

如果需要使用 shell 语法，请显式调用 shell：

```yaml
cmd: ["sh", "-c", "curl -s https://api.ipify.org | tr -d '\\n'"]
```

---

## `exitCode`

期望的命令退出码。其它退出码都会被视为错误。

```yaml
exitCode: 0
```

---

## `env`

追加到命令执行环境中的环境变量。会保留继承自 Lightddns 进程的环境变量，并追加这里配置的 `KEY=VALUE` 项。

```yaml
env:
  - PATH=/usr/bin:/bin:/usr/sbin:/sbin
  - TOKEN={{ .Env.MY_TOKEN }}
```

---

## `output`

控制哪些输出流转发到 Lightddns 自身的 stdout/stderr，便于观察命令执行情况。它不决定解析哪个输出流。

| 值 | 行为 |
|---|---|
| `none` 或留空 | 不转发命令输出。 |
| `stdout` | 转发 stdout。 |
| `stderr` | 转发 stderr。 |
| `all` | 同时转发 stdout 和 stderr。 |

```yaml
output: stderr
```

---

## `capture`

控制捕获并解析哪些输出流。

| 值 | 行为 |
|---|---|
| 留空 | 等同于 `stdout`。 |
| `stdout` | 解析 stdout。 |
| `stderr` | 解析 stderr。 |
| `all` | 同时解析 stdout 和 stderr。 |
| `none` | 对此数据源无效。 |

```yaml
capture: stdout
```

---

## `stdin`

把指定文件内容作为命令的标准输入。

相对路径会基于该数据源的有效工作目录解析：

1. 若配置了 `workDir`，使用 `workDir`。
2. 否则使用 `lightddns -D/--workdir` 设置的全局工作目录。
3. 若未显式传 `-D`，其默认值是 `.`，即进程当前工作目录。

绝对路径会按原样读取。若同时设置 `stdin` 和 `stdinContent`，优先使用 `stdin`。

```yaml
stdin: input.txt
```

---

## `stdinContent`

内联标准输入内容。设置了 `stdin` 时此项会被忽略。

```yaml
stdinContent: |
  query payload
```

---

## `sync`

设为 `true` 时，同一个数据源的并发 IP 查询会串行执行。命令会读写共享本地状态时可以开启。

```yaml
sync: true
```

---

## `workDir`

命令进程的工作目录，同时也是相对 `stdin` 路径的基准目录。

留空时使用 `lightddns -D/--workdir` 设置的全局工作目录。若配置为相对路径，会基于全局工作目录解析。绝对路径会按原样使用。

```yaml
workDir: scripts
```

例如执行 `lightddns -D /etc/lightddns run -c config.yaml` 时，上面的配置会让命令在 `/etc/lightddns/scripts` 下运行。

---

## `match`

可选的 IP 提取规则。参见 [MatchOption](../shared/match.md)。

```yaml
match:
  jq: ".ip"
  regex: "IP:\\s+(\\S+)"
```
