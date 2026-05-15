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

参见 [HTTP 数据源](datasource/http.md)。

## `providers`

服务提供者列表。每个 Provider 负责将 IP 地址更新到对应的 DNS 服务商。

## `domains`

域名记录列表。每个条目将域名、数据源和服务提供者绑定在一起，实现自动 DDNS 更新。

参见 [域名](domain.md)。

## `log`

全局日志配置。参见 [日志](log.md)。
