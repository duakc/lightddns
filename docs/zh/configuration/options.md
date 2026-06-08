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
  - domain: example.com
    provider: prov-cf
    datasource: data-http

services:
  - type: prometheus
    name: svc-metrics
    # ... Prometheus 导出器配置
```

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