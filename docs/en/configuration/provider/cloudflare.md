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

# ... ConnectOption
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

## `ConnectOption`

See [ConnectOption](../shared/connect.md).

## `HTTPOption`

See [HTTPOption](../shared/http.md).
