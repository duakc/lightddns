# 服务

与 DDNS 主循环并行运行的后台 HTTP 服务，全部为可选。

## Prometheus 导出器

以 Prometheus 格式导出 domain / provider / datasource / service 的指标。一旦启用，指标注册表就不再是 no-op，开始真正记录样本。

```yaml
log:
  level: info

services:
  - type: prometheus
    name: svc-metrics
    enabled: true
    listen: "127.0.0.1"
    port: 9001
    path: /metrics
```

抓取 `http://127.0.0.1:9001/metrics`。所有指标都以 `lightddns_<subsystem>_*` 为前缀（如 `lightddns_domain_update_success_total`、`lightddns_provider_request_total`）。

## IP 回显服务

返回调用方的 IP。适合搭建自定义的 HTTP 数据源，或者调试反向代理头的传递。按以下顺序识别真实 IP：`Cf-Connecting-IP` → `True-Client-IP` → `X-Real-IP` → `X-Forwarded-For`，否则使用 TCP 远端地址。

```yaml
services:
  - type: ipserver
    name: svc-ip
    enabled: true
    listen: "0.0.0.0"
    port: 9002
    path: /
    dump: false
```

```bash
$ curl http://localhost:9002/
203.0.113.5

$ curl http://localhost:9002/?format=json
{"ip":"203.0.113.5","is_bogon":false}

$ curl http://localhost:9002/?format=yaml
ip: 203.0.113.5
is_bogon: false
```

!!! note "`dump`"
    设为 `true` 后，每次请求 / 响应（含 headers 与 body）都会以 `debug` 级别打印日志。在多层代理调试"到底是哪个 header 携带了真实 IP"时非常有用。
