# DNSOption

DNS 解析配置。

**字符串形式**

```yaml
# 系统解析器
dns: system

# DNS over TLS，默认端口 853
dns: tls://8.8.8.8

# DNS over TLS，显式指定端口
dns: tls://8.8.8.8:853
```

**对象形式**

```yaml
dns:
  enabled: true
  type: tls
  server: 8.8.8.8
  port: 853
```

## `enabled`

是否启用自定义 DNS 解析。
!!! note "关于启用"
    字符串形式始终启用；对象形式下需显式设置为 true 才可启用。

## `type`

可用值:
- `system` 使用本地系统解析（默认使用）。
- `tls` 使用 [DNS Over TLS](https://en.wikipedia.org/wiki/DNS_over_TLS)

## `server`

上游 DNS 服务器地址。

## `port`

上游服务器端口。缺省时根据不同的 `type` 使用默认端口。

!!! note
    仅当目标地址为域名时 DNS 配置才会生效。若目标为 IP 地址，则跳过 DNS 解析。
