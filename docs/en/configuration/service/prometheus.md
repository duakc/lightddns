This document was translated from Chinese by AI.

# Prometheus

Exports internal Prometheus metrics over HTTP.

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

## `enabled`

Controls whether the service is enabled.

## `listen`

The listening address. An empty value means all interfaces (`0.0.0.0` and `::`).

## `port`

The TCP listening port. Optional; defaults to `9001` when omitted.

## `path`

The HTTP path served by the service. Defaults to `/metrics`.

---

## Metric Naming

All metrics use the prefix `lightddns_<subsystem>_*`, where `<subsystem>` is one of `domain`, `provider`, `datasource`, or `service`. Common examples:

| Metric                                       | Type      | Labels              |
|----------------------------------------------|-----------|---------------------|
| `lightddns_domain_update_success_total`      | counter   | `domain`            |
| `lightddns_domain_update_failure_total`      | counter   | `domain`            |
| `lightddns_domain_update_duration_seconds`   | histogram | `domain`            |
| `lightddns_provider_request_total`           | counter   | `name`, `operation` |
| `lightddns_provider_request_failure_total`   | counter   | `name`, `operation` |
| `lightddns_build_info`                       | gauge     | `version`, `branch` |
