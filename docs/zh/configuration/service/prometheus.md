# Prometheus

通过 HTTP 导出内部 Prometheus 指标。

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

控制是否启用该服务。

## `listen`

监听地址。留空代表"所有接口"（`0.0.0.0` 和 `::`）。

## `port`

监听的 TCP 端口，可省略；缺省时为 `9001`。

## `path`

提供服务的 HTTP 路径。默认 `/metrics`。

---

## 指标命名

所有指标都以 `lightddns_<subsystem>_*` 为前缀，`<subsystem>` 取自 `domain`、`provider`、`datasource`、`service`。常见示例：

| 指标                                       | 类型      | Labels              |
|--------------------------------------------|-----------|---------------------|
| `lightddns_domain_update_success_total`    | counter   | `domain`            |
| `lightddns_domain_update_failure_total`    | counter   | `domain`            |
| `lightddns_domain_update_duration_seconds` | histogram | `domain`            |
| `lightddns_provider_request_total`         | counter   | `name`、`operation` |
| `lightddns_provider_request_failure_total` | counter   | `name`、`operation` |
| `lightddns_build_info`                     | gauge     | `version`、`branch` |
