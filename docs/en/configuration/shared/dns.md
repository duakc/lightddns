This document was translated from Chinese by AI.

# DNSOption

DNS resolution configuration.

**String form**

```yaml
# System resolver
dns: system

# DNS over TLS, default port 853
dns: tls://8.8.8.8

# DNS over TLS, explicit port
dns: tls://8.8.8.8:853
```

**Object form**

```yaml
dns:
  enabled: true
  type: tls
  server: 8.8.8.8
  port: 853
```

## `enabled`

Whether custom DNS resolution is enabled.

!!! note "Enabling DNS"
    The string form is always enabled. In the object form, this must be explicitly set to `true` to enable it.

## `type`

Available values:

- `system` uses the local system resolver (the default).
- `tls` uses [DNS over TLS](https://en.wikipedia.org/wiki/DNS_over_TLS).

## `server`

The upstream DNS server address.

## `port`

The upstream server port. When omitted, the default port for the selected `type` is used.

!!! note
    DNS configuration only takes effect when the target address is a domain name. DNS resolution is skipped when the target is an IP address.
