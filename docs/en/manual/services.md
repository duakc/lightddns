# Services

Background HTTP services that run alongside the DDNS loop. None are required.

## Prometheus exporter

Exposes domain / provider / datasource / service metrics in Prometheus format. When enabled, the metrics registry stops being a no-op and starts collecting samples.

```yaml
log:
  level: info

services:
  - type: prometheus
    name: svc-metrics
    enabled: true
    listen: "127.0.0.1"
    port: 9090
    path: /metrics
```

Scrape `http://127.0.0.1:9090/metrics`. All metrics are prefixed `lightddns_<subsystem>_*` (e.g. `lightddns_domain_update_success_total`, `lightddns_provider_request_total`).

## IP echo server

Returns the caller's IP. Handy for building your own HTTP datasource, or for testing reverse-proxy header propagation. It honours `Cf-Connecting-IP`, `True-Client-IP`, `X-Real-IP`, and `X-Forwarded-For`, falling back to the connection's remote address.

```yaml
services:
  - type: ipserver
    name: svc-ip
    enabled: true
    listen: "0.0.0.0"
    port: 8080
    path: /
    dump: false
```

```bash
$ curl http://localhost:8080/
203.0.113.5

$ curl http://localhost:8080/?format=json
{"ip":"203.0.113.5","is_bogon":false}

$ curl http://localhost:8080/?format=yaml
ip: 203.0.113.5
is_bogon: false
```

!!! note "`dump`"
    Set `dump: true` to log every request/response (headers + body) at `debug` level. Useful when debugging which header carries the real client IP behind a proxy chain.