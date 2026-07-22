# DNSOption

DNS resolution configuration, shared by datasources and Service Providers. It lives under the `dns:` key of a datasource or provider and can be written as a short string or as a full object.

**String form** (recommended) — always enabled:

```yaml
# system resolver
dns: system

# DNS over TLS, default port 853
dns: tls://8.8.8.8

# DNS over TLS, explicit port
dns: tls://8.8.8.8:853
```

**Object form** — lets you toggle `enabled` while keeping the configuration:

```yaml
dns:
  enabled: true
  type: tls
  server: 8.8.8.8
  port: 853
```

## `enabled`

Whether custom DNS resolution is applied. The string form is always enabled; in the object form you set it explicitly.

## `type`

Resolver type. Accepts `system` (use the system resolver) or `tls` (DNS over TLS).

## `server`

Upstream DNS server address. Only used when `type` is `tls`.

## `port`

Upstream server port. Defaults to `853` for DNS over TLS when omitted.

!!! note
    DNS config only takes effect when the target is a domain name — it is skipped when the target is an IP address.