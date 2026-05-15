# Domain

Binds a domain name to a Service Provider and a datasource for automatic DNS updates.

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

A fully-qualified domain name. Must be a valid domain name (RFC 1035, RFC 3696).

```yaml
domain: www.example.com
```

## `enabled`

Enables or disables this domain record. Disabled domains are skipped entirely.

## `provider`

References a Service Provider by its `name`. The provider handles the actual DNS record updates.

```yaml
providers:
  - type: cloudflare
    name: cf-primary
    # ...

domains:
  - domain: example.com
    provider: cf-primary  # references the provider above
```

## `datasource`

References a datasource by its `name`. The datasource discovers the current public IP address.

```yaml
datasources:
  - type: http
    name: data-http
    # ...

domains:
  - domain: example.com
    datasource: data-http  # references the datasource above
```

## `ttl`

DNS record TTL in seconds. When unset, the Service Provider default is used.

```yaml
ttl: 300  # 5 minutes
```

## `ipv4` / `ipv6`

Control which IP address families to update for this domain. At least one must be enabled.

```yaml
# IPv4 only
ipv4: true
ipv6: false

# IPv6 only
ipv4: false
ipv6: true

# Dual-stack
ipv4: true
ipv6: true
```

## `interval`

Polling interval for checking IP changes. Accepts Go duration strings (e.g. `5m`, `1h`, `300s`).

```yaml
interval: 5m
```

## `timeout`

Timeout for each IP check and DNS update operation. Accepts Go duration strings.

```yaml
timeout: 30s
```
