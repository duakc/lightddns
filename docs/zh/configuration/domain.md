# 域名

将域名绑定到服务提供者和数据源，实现 DNS 记录的自动更新。

```yaml
# required
domain: example.com
provider: cf-primary
datasource: data-http

enabled: true
ttl: 300
ipv4: true
ipv6: false
interval: 5m
timeout: 30s
```

## `domain`

完全限定域名。必须是符合 RFC 1035 和 RFC 3696 标准的有效域名。

```yaml
domain: www.example.com
```

## `enabled`

是否启用该域名记录。禁用的域名会被完全跳过。

## `provider`

通过 `name` 引用一个服务提供者，负责 DNS 记录的实际更新。

```yaml
providers:
  - type: cloudflare
    name: cf-primary
    # ...

domains:
  - domain: example.com
    provider: cf-primary  # 引用上面定义的服务提供者
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
    datasource: data-http  # 引用上面定义的数据源
```

## `ttl`

DNS 记录的 TTL，单位秒。不设置时使用服务提供者默认值。

```yaml
ttl: 300  # 5 分钟
```

## `ipv4` / `ipv6`

控制该域名的 IP 地址族。至少需要启用一个。

```yaml
# 仅 IPv4
ipv4: true
ipv6: false

# 仅 IPv6
ipv4: false
ipv6: true

# 双栈
ipv4: true
ipv6: true
```

## `interval`

检查 IP 变更的轮询间隔。支持 Go duration 字符串格式（如 `5m`、`1h`、`300s`）。

```yaml
interval: 5m
```

## `timeout`

每次 IP 检查和 DNS 更新操作的超时时间。支持 Go duration 字符串格式。

```yaml
timeout: 30s
```
