# Cloudflare

最小可用的 Cloudflare 配置：通过 HTTP 服务获取公网 IPv4，并保持 A 记录同步。

```yaml
log:
  level: info

datasources:
  - type: http
    name: data-http
    url: https://api.ip.sb/ip
    dialStrategy: ipv4_only

providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-http
    interval: 30s
    timeout: 10s
    ttl: 60
    ipv4: true
```

Token 权限要求：目标 Zone 的 `Zone:Read` + `DNS:Edit`。在 [Cloudflare token 控制台](https://dash.cloudflare.com/profile/api-tokens) 创建。

## 双栈

去掉 `dialStrategy`，再补充一个 IPv6 端点，即可同时维护 A 与 AAAA 记录：

```yaml
datasources:
  - type: http
    name: data-http
    url:
      ipv4: https://api.ipify.org
      ipv6: https://api6.ipify.org

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-http
    ipv4: true
    ipv6: true
```

## 启用代理（橙云）

在 Provider 上设置 `proxy: true`，由它管理的记录都会启用 Cloudflare 的代理：

```yaml
providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"
    proxy: true
```