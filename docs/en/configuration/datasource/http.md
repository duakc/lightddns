This document was translated from Chinese by AI.

# HTTP

Obtains IP addresses through HTTP requests.

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

    - `prefer_ipv6` / `prefer_ipv4`: creates both IPv4 and IPv6 request contexts and merges their results.
    - `ipv4_only`: creates only an IPv4 request context.
    - `ipv6_only`: creates only an IPv6 request context.

    Each request context uses the same `url`, `method`, `headers`, and `match` settings, but its network dialer is fixed to `tcp4` or `tcp6`.

---

## `url`

The request URL used to obtain IP addresses. It must include an `http://` or `https://` scheme.

```yaml
url: https://api.ip.sb/ip
```

## `method`

The HTTP request method. Defaults to `GET` when empty.

Canonical values are `GET`, `POST`, `PUT`, `HEAD`, `DELETE`, `PATCH`, `CONNECT`, `OPTIONS`, `TRACE`, `BREW`, `PROPFIND`, and `WHEN`. Input is case-insensitive and is converted to uppercase when loaded.

## `headers`

Custom HTTP request headers. Each value can be a single string or an array. If `User-Agent` is not configured, Lightddns adds its default User-Agent.

```yaml
headers:
  User-Agent: Lightddns/stable
  Authorization: "Bearer xxxxx"
  X-Custom-Header:
    - value1
    - value2
```

## `match`

Optional IP extraction rules. See [MatchOption](../shared/match.md).

The HTTP datasource has one special rule: when the response `Content-Type` is JSON and `match.jq` is configured, it does not fall back to a regular expression or plain text. This prevents unrelated response text from being mistaken for a valid result when a JSON endpoint is broken.

## `connect`

See [ConnectOption](../shared/connect.md).

## `http`

See [HTTPOption](../shared/http.md).

## `dns`

See [DNSOption](../shared/dns.md).
