# Configuration File

Structure overview of the Lightddns configuration file. See sub-pages for detailed field references.

```yaml
log:
  level: info

datasources:
  - type: http
    name: data-http
    # ... HTTP datasource config

providers:
  - type: cloudflare
    name: prov-cf
    # ... Cloudflare provider config

domains:
  - domain: example.com
    provider: prov-cf
    datasource: data-http

services:
  - type: prometheus
    name: svc-metrics
    # ... Prometheus exporter config
```

## Working Directory

The CLI global flag `-D` / `--workdir` sets Lightddns' working directory. It defaults to `.`.

This value affects:

- Relative `run -c/--config` paths: `-c config.yaml` is read from the working directory.
- Relative `--env-file` paths: `--env-file secrets.env` is read from the working directory.
- Relative `log.output` paths: they are created under the working directory.
- `command` datasources: when `workDir` is empty, commands run in the working directory; relative `command.workDir` values are resolved under it; relative `command.stdin` paths are resolved under the command's effective working directory.

Absolute paths are used as-is.

```bash
lightddns -D /etc/lightddns run -c lightddns.yaml --env-file secrets.env
```

In this example, `lightddns.yaml`, `secrets.env`, and relative log/command paths are based on `/etc/lightddns`.

## `log`

Global logging configuration. See [Log](log.md).

## `datasources`

A list of datasources. Each datasource discovers the current public IP address of the host.

| Type | Description |
|---|---|
| [`http`](datasource/http.md) | Discover public IP via HTTP request. Supports JSON (jq) and regex extraction. |
| [`netlink`](datasource/netlink.md) | Read IP addresses from local network interfaces. |
| [`command`](datasource/command.md) | Run shell commands to discover IP addresses. |
| [`sum`](datasource/sum.md) | Merge IPs from multiple child datasources. |
| [`failover`](datasource/failover.md) | Query child datasources in priority order, failing over on error. |
| [`filter`](datasource/filter.md) | Filter IPs from child datasources with CIDR prefix rules. |

## `providers`

A list of Service Providers. Each provider pushes IP updates to the corresponding DNS service.

| Type | Description |
|---|---|
| [`cloudflare`](provider/cloudflare.md) | Update DNS records via the Cloudflare API. Supports A/AAAA records and proxy mode. |
| [`aliyun`](provider/aliyun.md) | Update DNS records via Aliyun DNS (alidns). |
| [`tencentcloud`](provider/tencentcloud.md) | Update DNS records via Tencent Cloud DNSPod. |

## `domains`

A list of domain records. Each entry binds a domain, provider, and datasource together for automatic DDNS updates.

See [Domain](domain.md).

## `services`

A list of background HTTP services. All are optional.

| Type | Description |
|---|---|
| [`prometheus`](service/prometheus.md) | Expose internal metrics in Prometheus format. |
| [`ipserver`](service/ipserver.md) | Echo the caller's public IP — useful for building your own HTTP datasource. |
