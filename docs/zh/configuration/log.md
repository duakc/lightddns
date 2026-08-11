# 日志

全局日志配置。

```yaml
level: info
disabled: false
output: ""
```

## `level`

日志级别：

可用值: `debug | info | warn | error | panic | fatal`。

留空使用 `info`

## `disabled`

设为 `true` 时禁用所有日志输出。

## `output`

日志输出文件路径。为空时输出到标准输出（stdout）。相对路径会创建在 `lightddns -D/--workdir` 指定的全局工作目录下。

```yaml
output: "/var/log/lightddns.log"
```
