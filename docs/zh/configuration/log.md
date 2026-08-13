# 日志

全局日志配置。

```yaml
level: info
disabled: false
output: "stdout"
```

## `level`

日志级别：

规范值：`debug | info | warn | warning |  error | dpanic | panic | fatal`。

> 大小写不敏感。

省略时默认使用 `info`。

## `disabled`

设为 `true` 时将禁用所有日志输出。

## `output`

`stdout` 将输出到标准输出内(默认值)。
`stderr` 将输出到标准错误输出内。

```yaml
output: "/var/log/lightddns.log"
```
