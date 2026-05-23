# Log

Global logging configuration.

```yaml
level: info
disabled: false
output: ""
```

## `level`

Log level. Uses [zap](https://github.com/uber-go/zap) levels:

`debug | info | warn | error | panic | fatal`.

When empty, defaults to `info`.

## `disabled`

Disables all logging output when set to `true`.

## `output`

Log output file path. When empty, logs are written to stdout.

```yaml
output: "/var/log/lightddns.log"
```
