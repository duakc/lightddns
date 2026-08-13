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
  - enabled: true
    domain: example.com
    provider: prov-cf
    datasource: data-http

services:
  - type: prometheus
    name: svc-metrics
    enabled: true
    # ... Prometheus exporter config
```

`datasources`, `providers`, and `services` are typed lists. If `name` is
omitted, Lightddns fills one in automatically when the config is loaded.

Fields marked `omitempty` in the `options` package can be omitted from YAML.
The following top-level fields and type-specific fields are required by their
respective entries:

The top-level `datasources`, `providers`, `domains`, and `services` keys are
required because they do not use `omitempty`; use `[]` when a list has no
entries. To start the runtime, at least one domain or service entry must have
`enabled: true`.

| Section | Required fields | Available `type` values |
|---|---|---|
| `log` | none | not typed |
| `datasources` | `type` and the fields listed below | `http`, `netlink`, `command`, `sum`, `failover`, `filter` |
| `providers` | `type` and provider credentials | `cloudflare`, `aliyun`, `tencentcloud` |
| `domains` | `enabled`, `domain` | not typed |
| `services` | `type`, `enabled` | `prometheus`, `ipserver` |

The current runtime still places conditions on these `omitempty` fields:

- `netlink` needs at least one of `ifName` or `ifIndex` to return addresses.
- A domain may omit `provider` or `datasource` only when the corresponding
  manager contains exactly one item; otherwise use its `name` explicitly.
- An enabled DNS object with `type: tls` needs a non-empty `server`.
- If a domain sets `interval` below the default `timeout` of `15s`, it must also
  set `timeout` to a value no greater than `interval`.

The service ports are optional. When omitted, `prometheus` listens on `9001`
and `ipserver` listens on `9002`.

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

The `command.output` values are `none`, `stdout`, `stderr`, and `all`.
The `command.capture` values are `stdout`, `stderr`, and `all`; an omitted
`capture` defaults to `stdout`.

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
