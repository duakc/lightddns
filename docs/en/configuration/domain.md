This document was translated from Chinese by AI.

# Domain

Binds a domain to a service provider and datasource.

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

Whether this domain record is enabled.

## `domain`

A fully qualified domain name. It must be a valid domain name that complies with [RFC 1035](https://datatracker.ietf.org/doc/html/rfc1035) and [RFC 3696](https://datatracker.ietf.org/doc/html/rfc3696).

```yaml
domain: www.example.com
```

## `provider`

??? note "Simplified configuration"
    This field may be omitted when only one `provider` is configured globally.

References a service provider by `name`. The provider performs the DNS record updates.

```yaml
providers:
  - type: cloudflare
    name: prov-main
    # ...

domains:
  - enabled: true
    domain: example.com
    provider: prov-main
    # ...
```

## `datasource`

??? note "Simplified configuration"
    This field may be omitted when only one `datasource` is configured globally.

References a datasource by `name`. The datasource obtains the current public IP address.

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

The DNS record TTL in seconds.

!!! note
    If set to `0`, different providers may use different default TTL values.
    The resulting value depends on the provider implementation.

```yaml
ttl: 300
```

## `ipv4` / `ipv6`

Controls which IP versions (IPv4 / IPv6) are updated for this domain.
Leaving them empty enables both IPv4 and IPv6.

## `interval`

The polling interval for checking IP changes. Supports Go duration strings such as `5m`, `1h`, and `300s`. Optional; defaults to `30s`.

```yaml
interval: 5m
```

## `timeout`

The timeout for each IP check and DNS update operation. Supports Go duration strings.
Defaults to `15s` and cannot be greater than `interval`.
If `interval` is less than `15s`, explicitly set `timeout` to a value no greater than that interval.

```yaml
timeout: 30s
```
