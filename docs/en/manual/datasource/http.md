This document was translated from Chinese by AI.

## Common Services

The following configurations cover several common public IP services. Service availability, rate limits, and the address families actually returned are determined by the external services themselves.

### Plain Text

When the response body contains only an IP address, `match` can be omitted.

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
  url: https://api.ip.sb/jsonip
  match:
    jq: ".ip"
```

### Text with Explanatory Content

Use `match.regex` when the page includes explanatory content in addition to the IP address.

```yaml
- type: http
  name: ipip
  url: https://myip.ipip.net
  match:
    regex: "当前 IP：\\s*(.+?)\\s*来自于："
```
