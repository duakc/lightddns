# HTTP

Retrieves public IP addresses through an HTTP request. It supports plain-text responses, JSON extraction through `match.jq`, regex extraction through `match.regex`, and IPv4/IPv6 connection selection.

```yaml
# required
type: http
name: data-http
url: https://api.ip.sb/ip

# optional
method: GET
headers:
  User-Agent: Lightddns/stable
match:
  jq: ""
  regex: ""

dns: system
connect:
  # ... ConnectOption
http:
  # ... HTTPOption
```

??? note "Behavior"
    The HTTP datasource creates request contexts according to `connect.dialStrategy`:

    - `prefer_ipv6` / `prefer_ipv4`: create both IPv4 and IPv6 request contexts, then merge results.
    - `ipv4_only`: create only the IPv4 request context.
    - `ipv6_only`: create only the IPv6 request context.

    Each request context uses the same `url`, `method`, `headers`, and `match` settings, but its network dialer is pinned to `tcp4` or `tcp6`.

---

## Common Services

These examples show the expected config shape for common public IP endpoints. Provider availability, rate limits, and returned address family are controlled by those external services.

### Plain Text

Use no `match` block when the response body is only an IP address.

```yaml
- type: http
  name: ip-sb
  url: https://api.ip.sb/ip
```

```yaml
- type: http
  name: ipify-v4
  url: https://api.ipify.org
  connect:
    dialStrategy: ipv4_only
```

```yaml
- type: http
  name: ipify-v6
  url: https://api6.ipify.org
  connect:
    dialStrategy: ipv6_only
```

### JSON

Use `match.jq` for JSON endpoints.

```yaml
- type: http
  name: ipinfo
  url: https://ipinfo.io
  match:
    jq: ".ip"
```

```yaml
- type: http
  name: ipify-json
  url: https://api64.ipify.org?format=json
  match:
    jq: ".ip"
```

### Text With Labels

Use `match.regex` when the page includes text around the address.

```yaml
- type: http
  name: ipip
  url: https://myip.ipip.net
  match:
    regex: "当前 IP：\\s*(.+?)\\s*来自于："
```

---

## `url`

The request URL for retrieving the IP address. It must include the `http://` or `https://` scheme.

```yaml
url: https://api.ip.sb/ip
```

---

## `method`

HTTP request method. Empty defaults to `GET`.

Supported values: `GET`, `POST`, `PUT`, `HEAD`, `DELETE`, `PATCH`, `CONNECT`, `OPTIONS`, `TRACE`, plus `BREW`, `PROPFIND`, `WHEN`.

---

## `headers`

Custom HTTP request headers. Values can be a single string or an array. If `User-Agent` is omitted, Lightddns adds its default User-Agent.

```yaml
headers:
  User-Agent: Lightddns/stable
  Authorization: "Bearer xxxxx"
  X-Custom-Header:
    - value1
    - value2
```

---

## `match`

Optional extraction rules. See [MatchOption](../shared/match.md).

For HTTP specifically, JSON `Content-Type` plus `match.jq` disables fallback to regex/plain text. This prevents a broken JSON endpoint from being accidentally accepted through unrelated response text.

---

## `connect`

See [ConnectOption](../shared/connect.md).

## `http`

See [HTTPOption](../shared/http.md).

## `dns`

See [DNSOption](../shared/dns.md).
