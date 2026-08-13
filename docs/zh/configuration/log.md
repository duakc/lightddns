# 日志

全局日志配置。

```yaml
level: info
disabled: false
output: ""
```

## `level`

日志级别：

规范值：`debug | info | warn | error | dpanic | panic | fatal`。大小写不
敏感，`warning` 也可作为 `warn` 的别名。

省略时默认使用 `info`。

## `disabled`

设为 `true` 时禁用所有日志输出。

## `output`

日志输出目标。特殊值为 `stdout`（留空时的默认值）和 `stderr`；其它值
都会被当作文件路径。相对路径会创建在 `lightddns -D/--workdir` 指定的
全局工作目录下。

```yaml
output: "/var/log/lightddns.log"
```
