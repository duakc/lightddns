# HTTP Datasource

Retrieves the public IP address via HTTP(S). Supports JSON (jq) extraction, regex extraction, and dual-stack networking (IPv4/IPv6).

## Example

```yaml
# required
type: http
name: data-http
url: https://api64.ipify.org?format=json

json:
  ipv4: ".ip"
  ipv6: ".ip"
regex:
  ipv4: ""
  ipv6: ""
method: GET
headers:
  User-Agent: Lightddns/stable

# ... ConnectOption
# ... HTTPOption
```

---

## `url`

The request URL for retrieving the IP address. **Must include the `http://` or `https://` scheme**.

```yaml
# Returns plain-text IP
url: https://api.ip.sb/ip

# Returns JSON
url: https://api64.ipify.org?format=json
```

If the URL host is an IP address rather than a domain name, `dialStrategy` is automatically overridden to match the IP's address family (IPv4 → `ipv4_only`, IPv6 → `ipv6_only`).

---

## `json`

Extracts IP addresses from JSON responses using [jq](https://github.com/itchyny/gojq) syntax.

**Short form** (`string`) — same jq expression for both IPv4 and IPv6:

```yaml
json: ".ip"
```

**Object form** (`DualStack<string>`) — separate jq expressions per address family:

```yaml
json:
  ipv4: ".ipv4"
  ipv6: ".ipv6"
```

When the response `Content-Type` is `application/json` and `json` is configured, JSON path extraction takes priority. Each jq result value is parsed as an IP address; ensure the returned value is a valid IPv4 or IPv6 string.

---

## `regex`

Extracts IP addresses from non-JSON responses using a regular expression. At least one of `json` or `regex` must be configured.

**Short form** (`string`) — same regex for both IPv4 and IPv6:

```yaml
regex: "IP:\\s*(.+?)\\n"
```

**Object form** (`DualStack<string>`) — separate regex per address family:

```yaml
regex:
  ipv4: "IPv4:\\s*(.+?)\\n"
  ipv6: "IPv6:\\s*(.+?)\\n"
```

Uses `FindAllSubmatch`; captures the **first group** as the IP address.

!!! note "Extraction priority"
    1. If `Content-Type` is `application/json` and `json` is set → JSON (jq) extraction
    2. If `regex` is set → regex extraction
    3. Otherwise → treat response body as plain-text IP

---

## `method`

HTTP request method. Supports `GET`, `POST`, `PUT`, `HEAD`, `DELETE`, `PATCH`, `CONNECT`, `OPTIONS`, `TRACE`, plus `BREW`, `PROPFIND`, `WHEN`.

---

## `headers`

Custom HTTP request headers. Values can be a single string or an array.

```yaml
headers:
  User-Agent: Lightddns/stable
  Authorization: "Bearer xxxxx"
  X-Custom-Header:
    - value1
    - value2
```

---

## Connection Options

See [ConnectOption](../shared/Connect.md).

## HTTP Client Options

See [HTTPOption](../shared/HTTP.md).

---

## Dual-Stack Behavior

The HTTP datasource implements the `DatasourceDualStack` interface (exposing both `IPv4()` and `IPv6()`). It creates separate IPv4 and IPv6 request contexts based on `dialStrategy`:

- Non-`ipv4_only` mode → creates IPv6 context (`tcp6` dialer)
- Non-`ipv6_only` mode → creates IPv4 context (`tcp4` dialer)

Each context uses its own `json`/`regex` expression for IP extraction. The `IP()` method merges both results.
