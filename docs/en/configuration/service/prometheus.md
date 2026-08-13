# Prometheus

Exposes Lightddns's internal metrics over HTTP in Prometheus format.

```yaml
# required
type: prometheus
name: svc-metrics
enabled: true

# optional
listen: ""
port: 9001
path: /metrics
```

When no Prometheus service is enabled the metrics registry is a no-op and no samples are collected.

## `enabled`

Must be `true` for the service to start. When `false`, this entry is silently skipped.

## `listen`

Bind address. Empty means "all interfaces" (`0.0.0.0` + `::`). Use `127.0.0.1` to expose metrics only on loopback.

## `port`

TCP port to listen on. Optional; defaults to `9001` when omitted.

## `path`

HTTP path served. Defaults to `/metrics`.

---

## Metric naming

All metrics are prefixed `lightddns_<subsystem>_*`, where `<subsystem>` is one of `domain`, `provider`, `datasource`, or `service`. A few examples:

| Metric | Type | Labels |
|---|---|---|
| `lightddns_domain_update_success_total` | counter | `domain` |
| `lightddns_domain_update_failure_total` | counter | `domain` |
| `lightddns_domain_update_duration_seconds` | histogram | `domain` |
| `lightddns_provider_request_total` | counter | `name`, `operation` |
| `lightddns_provider_request_failure_total` | counter | `name`, `operation` |
| `lightddns_build_info` | gauge | `version`, `branch` |
