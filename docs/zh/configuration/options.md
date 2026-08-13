# 配置文件

Lightddns 配置文件的结构概览。完整示例可参考各子章节。

```yaml
log:
  level: info

datasources:
  - type: http
    name: data-http
    # ... HTTP 数据源配置

providers:
  - type: cloudflare
    name: prov-cf
    # ... Cloudflare Provider 配置

domains:
  - enabled: true
    domain: example.com
    provider: prov-cf
    datasource: data-http

services:
  - type: prometheus
    name: svc-metrics
    enabled: true
    # ... Prometheus 导出器配置
```

`datasources`、`providers` 和 `services` 都是带 `type` 的列表。如果
`name` 留空，Lightddns 在加载配置时会自动补上。

`options` 包中标记了 `omitempty` 的字段可以从 YAML 中省略。各顶层
配置段及具体类型字段的必填关系如下：

顶层的 `datasources`、`providers`、`domains`、`services` 没有使用
`omitempty`，因此这些键必须存在；没有条目时写成 `[]`。程序要实际启动，
至少需要一个设置了 `enabled: true` 的域名或服务。

| 配置段 | 必填字段 | `type` 可选值 |
|---|---|---|
| `log` | 无 | 无类型 |
| `datasources` | `type` 及下方各数据源要求的字段 | `http`、`netlink`、`command`、`sum`、`failover`、`filter` |
| `providers` | `type` 及 Provider 凭据 | `cloudflare`、`aliyun`、`tencentcloud` |
| `domains` | `enabled`、`domain` | 无类型 |
| `services` | `type`、`enabled` | `prometheus`、`ipserver` |

当前实现仍对以下标记了 `omitempty` 的字段附加了运行时条件：

- `netlink` 至少需要填写 `ifName` 或 `ifIndex`，否则无法返回地址。
- 只有在对应管理器中恰好配置了一个条目时，域名才可以省略 `provider`
  或 `datasource`；多个条目时必须明确填写其 `name`。
- 启用 DNS 且使用 `type: tls` 时，`server` 不能为空。
- 若域名的 `interval` 小于默认 `timeout` `15s`，还必须显式设置不大于
  `interval` 的 `timeout`。

服务端口可以省略：`prometheus` 默认监听 `9001`，`ipserver` 默认监听
`9002`。

## 工作目录

CLI 全局参数 `-D` / `--workdir` 用于设置 Lightddns 的工作目录，默认值是 `.`。

这个值会影响：

- 相对 `run -c/--config` 路径：`-c config.yaml` 会从工作目录读取。
- 相对 `--env-file` 路径：`--env-file secrets.env` 会从工作目录读取。
- 相对 `log.output` 路径：日志文件会创建在工作目录下。
- `command` 数据源：`workDir` 留空时，命令在工作目录中执行；相对 `command.workDir` 会基于工作目录解析；相对 `command.stdin` 会基于命令的有效工作目录解析。

绝对路径会按原样使用。

```bash
lightddns -D /etc/lightddns run -c lightddns.yaml --env-file secrets.env
```

在这个例子中，`lightddns.yaml`、`secrets.env` 以及相对日志/命令路径都基于 `/etc/lightddns`。

## `log`

全局日志配置。参见 [日志](log.md)。

## `datasources`

数据源列表。每个数据源负责获取当前主机的公网 IP 地址。

| Type | 说明 |
|---|---|
| [`http`](datasource/http.md) | 通过 HTTP 请求获取公网 IP。支持 JSON（jq）和正则提取。 |
| [`netlink`](datasource/netlink.md) | 从本地网络接口读取 IP 地址。 |
| [`command`](datasource/command.md) | 通过执行 Shell 命令获取 IP 地址。 |
| [`sum`](datasource/sum.md) | 合并多个子数据源的 IP 地址。 |
| [`failover`](datasource/failover.md) | 按优先级顺序查询子数据源，失败时自动切换。 |
| [`filter`](datasource/filter.md) | 使用 CIDR 前缀规则过滤子数据源返回的 IP 地址。 |

`command.output` 可用值为 `none`、`stdout`、`stderr`、`all`。
`command.capture` 可用值为 `stdout`、`stderr`、`all`；省略 `capture` 时默认使用
`stdout`。

## `providers`

服务提供方列表。每个 Provider 负责将 IP 地址更新到对应的 DNS 服务商。

| Type | 说明 |
|---|---|
| [`cloudflare`](provider/cloudflare.md) | 通过 Cloudflare API 更新 DNS 记录。支持 A/AAAA 记录和代理模式。 |
| [`aliyun`](provider/aliyun.md) | 通过阿里云解析（alidns）更新 DNS 记录。 |
| [`tencentcloud`](provider/tencentcloud.md) | 通过腾讯云 DNSPod 更新 DNS 记录。 |

## `domains`

域名记录列表。每个条目将域名、数据源和服务提供方绑定在一起，实现自动 DDNS 更新。

参见 [域名](domain.md)。

## `services`

后台 HTTP 服务列表。全部为可选项。

| Type | 说明 |
|---|---|
| [`prometheus`](service/prometheus.md) | 以 Prometheus 格式导出内部指标。 |
| [`ipserver`](service/ipserver.md) | 回显调用方的公网 IP — 用于搭建自定义的 HTTP 数据源等。 |
