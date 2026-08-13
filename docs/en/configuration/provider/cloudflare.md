This document was translated from Chinese by AI.

# Cloudflare

Updates DNS records through the [Cloudflare API](https://developers.cloudflare.com/api/).

```yaml
# required
type: cloudflare
name: prov-cf
token: "your-cloudflare-api-token"

# optional
proxy: false

dns: system
connect:
  # ... ConnectOption
http:
  # ... HTTPOption
```

## `token`

Cloudflare API token. Create it in the [Cloudflare dashboard](https://dash.cloudflare.com/profile/api-tokens) with the following permissions:

- **Zone**: `DNS:Edit`
- **Zone**: `Zone:Read`

```yaml
token: "your-cloudflare-api-token"
```

---

## `proxy`

Whether to enable the Cloudflare proxy (orange cloud).

```yaml
proxy: true
```

---

## `connect`

See [ConnectOption](../shared/connect.md).

## `http`

See [HTTPOption](../shared/http.md).

## `dns`

See [DNSOption](../shared/dns.md).
