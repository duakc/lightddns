# Cloudflare

Updates DNS records via the [Cloudflare API](https://developers.cloudflare.com/api/).
Supports A and AAAA records with optional proxy (orange cloud) mode.

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

Cloudflare API token. Create one in the [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens) with the following permissions:

- **Zone** — `DNS:Edit`
- **Zone** — `Zone:Read`

```yaml
token: "your-cloudflare-api-token"
```

---

## `proxy`

Enables Cloudflare's proxy (orange cloud) for DNS records.

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
