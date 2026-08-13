This document was translated from Chinese by AI.

# Log

Global logging configuration.

```yaml
level: info
disabled: false
output: "stdout"
```

## `level`

Log level:

Canonical values: `debug | info | warn | warning | error | dpanic | panic | fatal`.

> Values are case-insensitive.

Defaults to `info` when omitted.

## `disabled`

Setting this to `true` disables all log output.

## `output`

`stdout` writes to standard output (the default).
`stderr` writes to standard error.

```yaml
output: "/var/log/lightddns.log"
```
