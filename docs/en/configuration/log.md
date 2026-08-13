# Log

Global logging configuration.

```yaml
level: info
disabled: false
output: ""
```

## `level`

Log level. Uses [zap](https://github.com/uber-go/zap) levels:

Canonical values: `debug | info | warn | error | dpanic | panic | fatal`.
They are case-insensitive, and `warning` is accepted as an alias for `warn`.

When omitted, defaults to `info`.

## `disabled`

Disables all logging output when set to `true`.

## `output`

Log output destination. The special values are `stdout` (the default when
empty) and `stderr`; any other value is treated as a file path. Relative paths
are created under the global working directory from `lightddns -D/--workdir`.

```yaml
output: "/var/log/lightddns.log"
```
