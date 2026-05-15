# 日志

全局日志配置。

```yaml
log:
  level: info

  disabled: false
  output: ""
```

## `level`

日志级别。使用 [zap](https://github.com/uber-go/zap) 日志级别：`debug | info | warn | error | dpanic | panic | fatal`。

## `disabled`

设为 `true` 时禁用所有日志输出。

## `output`

日志输出文件路径。为空时输出到标准输出（stdout）。

```yaml
output: "/var/log/lightddns.log"
```
