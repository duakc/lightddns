# DNSOption

DNS 解析配置，被数据源和服务提供者复用。它位于数据源或服务提供者的 `dns:` 键下，可以写成简短的字符串，也可以写成完整的对象。

**字符串形式**（推荐）— 始终启用：

```yaml
# 系统解析器
dns: system

# DNS over TLS，默认端口 853
dns: tls://8.8.8.8

# DNS over TLS，显式指定端口
dns: tls://8.8.8.8:853
```

**对象形式** — 可在保留配置的同时通过 `enabled` 开关启停：

```yaml
dns:
  enabled: true
  type: tls
  server: 8.8.8.8
  port: 853
```

## `enabled`

是否启用自定义 DNS 解析。字符串形式始终启用；对象形式下需显式设置。

## `type`

解析器类型枚举：`system`（使用系统解析器）或 `tls`（DNS over TLS）。

## `server`

上游 DNS 服务器地址。启用 DNS 且 `type` 为 `tls` 时必填，其它情况下不使用。

## `port`

上游服务器端口。使用 DNS over TLS 且省略时默认为 `853`。

!!! note
    仅当目标地址为域名时 DNS 配置才会生效。若目标为 IP 地址，则跳过 DNS 解析。
