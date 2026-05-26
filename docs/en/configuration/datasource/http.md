# HTTP

Retrieves the public IP address via HTTP. Supports JSON (jq) extraction, regex extraction, and dual-stack networking (IPv4/IPv6).

```yaml
# required
type: http
name: data-http
url: https://api64.ipify.org?format=json

# optional
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

??? note "Behavior"
    The HTTP datasource creates independent IPv4 and IPv6 request contexts, each pinned to its own `tcp4`/`tcp6` dialer:

    - An IPv4 context is created when `url.ipv4` is set and `dialStrategy` is not `ipv6_only`
    - An IPv6 context is created when `url.ipv6` is set and `dialStrategy` is not `ipv4_only`

    Each context uses its own URL plus the corresponding `json`/`regex` expression. The final result merges IPv4 and IPv6.

---

[JQ and Regex Examples](../example/http.md)

## `url`

The request URL for retrieving the IP address. **Must include the `http://` or `https://` scheme**.

**Short form** (`string`) — same URL used for both IPv4 and IPv6 requests:

```yaml
# Returns plain-text IP
url: https://api.ip.sb/ip

# Returns JSON
url: https://api64.ipify.org?format=json
```

**Object form** — separate endpoints for each stack. Useful when IPv4 and IPv6 lookups live on different hosts. Either stack may be omitted to disable that family entirely.

```yaml
url:
  ipv4: https://api.ipify.org
  ipv6: https://api6.ipify.org
```

If a URL host is a literal IP address rather than a domain name, the address family must match the stack it is assigned to (an IPv4 literal under `ipv4`, an IPv6 literal under `ipv6`).

---

## `json`

Extracts IP addresses from JSON responses using [jq](https://github.com/itchyny/gojq) syntax.

**Short form** (`string`) — same jq expression for both IPv4 and IPv6:

```yaml
json: ".ip"
```

**Object form** — separate jq expressions for IPv4 and IPv6:

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

**Object form** — separate regex for IPv4 and IPv6:

```yaml
regex:
  ipv4: "IPv4:\\s*(.+?)\\n"
  ipv6: "IPv6:\\s*(.+?)\\n"
```

Takes the **first match** as the IP address.

!!! note "Extraction priority"
    1. If `Content-Type` is `application/json` and `json` is set — JSON (jq) extraction
    2. If `regex` is set — regex extraction
    3. Otherwise — treat response body as plain-text IP

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

## `ConnectOption`

See [ConnectOption](../shared/connect.md).

## `HTTPOption`

See [HTTPOption](../shared/http.md).
