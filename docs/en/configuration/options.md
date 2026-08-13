This document was translated from Chinese by AI.

# Configuration File

An overview of the Lightddns configuration file structure. See the individual subsections for complete examples.

```yaml
log:
  level: info

datasources:
  - type: http
    name: data-http
    # ... HTTP datasource configuration

providers:
  - type: cloudflare
    name: prov-cf
    # ... Cloudflare provider configuration

domains:
  - enabled: true
    domain: example.com
    provider: prov-cf
    datasource: data-http

services:
  - type: prometheus
    name: svc-metrics
    enabled: true
    # ... Prometheus exporter configuration
```

`datasources`, `providers`, and `services` are lists whose entries have a `type`. If `name` is empty, Lightddns automatically fills it in while loading the configuration.

??? "Automatic `name` completion rules"
    ```yaml
    datasources:
        - type: http
        - type: command
    providers:
        - type: cloudflare
        - type: aliyun
    services:
        - type: ipserver
        - type: prometheus
    ```
    After automatic completion, this becomes:
    ```yaml
    datasources:
        - type: http
          name: datasource[0]
        - type: command
          name: datasource[1]
    providers:
        - type: cloudflare
          name: provider[0]
        - type: aliyun
          name: provider[1]
    services:
        - type: ipserver
          name: service[0]
        - type: prometheus
          name: service[1]
    ```

## Env File

The global CLI option `--env-file` specifies an additional source of environment variables for Lightddns. The file must follow the [standard dotenv syntax](https://genkitlab.com/blog/dotenv-file-syntax/).

```env
# comment is allowed
PROVIDER_TOKEN=this_is_some_random_token_here
```

## Working Directory

The global CLI options `-D` / `--workdir` set the Lightddns working directory. The default is the directory from which the user runs the program.
Setting the working directory affects how all relative paths in the program are converted to absolute paths.
Absolute paths are used as-is and are not affected by `-D` / `--workdir`.
For example:

```bash
$ lightddns -D /path_to_your_dir/lightddns -c config.yaml
```

This reads `/path_to_your_dir/lightddns/config.yaml`.

```bash
lightddns -D /etc/lightddns --env-file secrets.env run -c /etc/lightddns.yaml
```

This reads `/etc/lightddns/secrets.env` and `/etc/lightddns.yaml`.

## `log`

Global logging configuration. See [Log](log.md).

## `datasources`

A list of datasources. Each datasource obtains the current IP address of the host.

| Type                                 | Description                                                        |
|--------------------------------------|--------------------------------------------------------------------|
| [`http`](datasource/http.md)         | Obtains IP addresses through HTTP requests. Supports JSON (jq) and regular-expression extraction. |
| [`netlink`](datasource/netlink.md)   | Reads IP addresses from local network interfaces.                  |
| [`command`](datasource/command.md)   | Obtains IP addresses by running shell commands.                    |
| [`sum`](datasource/sum.md)           | Merges IP addresses from multiple child datasources.               |
| [`failover`](datasource/failover.md) | Queries child datasources in priority order and automatically switches on failure. |
| [`filter`](datasource/filter.md)     | Filters IP addresses returned by child datasources using CIDR prefix rules. |

## `providers`

A list of service providers. Each provider compares IP addresses and updates them at the corresponding DNS provider.

| Type                                       | Description                              |
|--------------------------------------------|------------------------------------------|
| [`cloudflare`](provider/cloudflare.md)     | Updates through the Cloudflare API.      |
| [`aliyun`](provider/aliyun.md)             | Updates through Aliyun DNS (alidns).     |
| [`tencentcloud`](provider/tencentcloud.md) | Updates through Tencent Cloud DNSPod.    |

## `domains`

A list of domain records. Each entry binds a domain, datasource, and service provider to perform automatic DDNS updates.

See [Domain](domain.md).

## `services`

A list of background HTTP services. All entries are optional.

| Type                                  | Description                         |
|---------------------------------------|-------------------------------------|
| [`prometheus`](service/prometheus.md) | Exports internal Prometheus metrics. |
| [`ipserver`](service/ipserver.md)     | A lightweight IP echo server.       |
