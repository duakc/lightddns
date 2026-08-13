# Domain

Binds a domain name to a Service Provider and a datasource.

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

Enables or disables this domain record.

## `domain`

A fully-qualified domain name. Must be a valid domain name per [RFC 1035](https://datatracker.ietf.org/doc/html/rfc1035) and [RFC 3696](https://datatracker.ietf.org/doc/html/rfc3696).

```yaml
domain: www.example.com
```

## `provider`

References a Service Provider by its `name`. The provider handles the actual DNS record updates.

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

References a datasource by its `name`. The datasource discovers the current public IP address.

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

DNS record TTL in seconds. When unset, the provider default is used.

!!! note
    Setting TTL to 0 may result in a provider-specific default TTL.

```yaml
ttl: 300
```

## `ipv4` / `ipv6`

Control which IP versions (IPv4 / IPv6) to update for this domain.
When both are left unset, IPv4 and IPv6 are both enabled.

## `interval`

Polling interval for checking IP changes. Accepts Go duration strings (e.g.
`5m`, `1h`, `300s`). Optional; defaults to `30s`.

```yaml
interval: 5m
```

## `timeout`

Timeout for each IP check and DNS update operation. Accepts Go duration strings.
Optional; defaults to `15s`. It must not be greater than `interval`.
When `interval` is less than `15s`, set `timeout` explicitly to a value no
greater than that interval.

```yaml
timeout: 30s
```
