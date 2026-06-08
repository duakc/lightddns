# Prometheus

通过 HTTP 以 Prometheus 格式导出 Lightddns 内部指标。

```yaml
# required
type: prometheus
name: svc-metrics
enabled: true

# optional
listen: ""
port: 9090
path: /metrics
```

如果没有启用任何 Prometheus 服务，指标注册表就是 no-op，不会真正记录任何样本。

## `enabled`

必须为 `true` 才会启动。`false` 时该条目会被静默跳过。

## `listen`

监听地址。留空代表"所有接口"（`0.0.0.0` + `::`）。设为 `127.0.0.1` 可仅在 loopback 暴露。

## `port`

监听的 TCP 端口。默认 `9090`。

## `path`

提供服务的 HTTP 路径。默认 `/metrics`。

---

## 指标命名

所有指标都以 `lightddns_<subsystem>_*` 为前缀，`<subsystem>` 取自 `domain`、`provider`、`datasource`、`service`。常见示例：

| 指标 | 类型 | Labels |
|---|---|---|
| `lightddns_domain_update_success_total` | counter | `domain` |
| `lightddns_domain_update_failure_total` | counter | `domain` |
| `lightddns_domain_update_duration_seconds` | histogram | `domain` |
| `lightddns_provider_request_total` | counter | `name`、`operation` |
| `lightddns_provider_request_failure_total` | counter | `name`、`operation` |
| `lightddns_build_info` | gauge | `version`、`branch` |