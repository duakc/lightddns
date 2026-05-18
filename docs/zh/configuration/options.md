# 配置文件

Lightddns 配置文件的结构概览。完整示例可参考各子章节。

```yaml
# required
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

log:
  level: info
```

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

服务提供者列表。每个 Provider 负责将 IP 地址更新到对应的 DNS 服务商。

| Type | 说明 |
|---|---|
| [`cloudflare`](provider/cloudflare.md) | 通过 Cloudflare API 更新 DNS 记录。支持 A/AAAA 记录和代理模式。 |

## `domains`

域名记录列表。每个条目将域名、数据源和服务提供者绑定在一起，实现自动 DDNS 更新。

参见 [域名](domain.md)。

## `log`

全局日志配置。参见 [日志](log.md)。
