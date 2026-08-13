# 域名

将域名绑定到服务提供者和数据源。

```yaml
# required
enabled: true
domain: example.com
provider: prov-cf
datasource: data-http

# optional
ttl: 300
ipv4: true
ipv6: false
interval: 5m
timeout: 30s
```

## `enabled`

是否启用该域名记录。

## `domain`

完全限定域名。必须是符合 [RFC 1035](https://datatracker.ietf.org/doc/html/rfc1035) 和 [RFC 3696](https://datatracker.ietf.org/doc/html/rfc3696) 标准的有效域名。

```yaml
domain: www.example.com
```

## `provider`

通过 `name` 引用一个服务提供者，负责 DNS 记录的实际更新。

```yaml
providers:
  - type: cloudflare
    name: prov-cf
    # ...

domains:
  - domain: example.com
    provider: prov-cf
```

## `datasource`

通过 `name` 引用一个数据源，负责获取当前公网 IP 地址。

```yaml
datasources:
  - type: http
    name: data-http
    # ...

domains:
  - domain: example.com
    datasource: data-http
```

## `ttl`

DNS 记录的 TTL，单位秒。 不设置时使用 provider 默认值。

!!! note
    如果设置为0，根据不同的 provider 可能会有不同的默认TTL。

```yaml
ttl: 300
```

## `ipv4` / `ipv6`

控制该域名要更新的 IP 版本（IPv4 / IPv6）。
留空代表 IPv4 和 IPv6 同时启用。


## `interval`

检查 IP 变更的轮询间隔。支持 Go duration 字符串格式（如 `5m`、`1h`、
`300s`）。可省略，默认 `30s`。

```yaml
interval: 5m
```

## `timeout`

每次 IP 检查和 DNS 更新操作的超时时间。支持 Go duration 字符串格式。
可省略，默认 `15s`，且不能大于 `interval`。
如果 `interval` 小于 `15s`，必须显式设置一个不大于该间隔的 `timeout`。

```yaml
timeout: 30s
```
