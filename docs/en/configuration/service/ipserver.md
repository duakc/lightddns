# IPServer

A tiny HTTP server that returns the caller's IP address. Useful as the upstream for an HTTP datasource, or for debugging which header your reverse proxy is forwarding.

```yaml
# required
type: ipserver
name: svc-ip
enabled: true
port: 8080

# optional
listen: ""
path: /
dump: false
```

??? note "Behavior"
    The server inspects request headers in this order and returns the first valid IP it finds: `Cf-Connecting-IP`, `True-Client-IP`, `X-Real-IP`, `X-Forwarded-For`. If none are present (or none parse as an IP), the TCP remote address is used.

    Comma-separated header values (typical for `X-Forwarded-For`) take the first entry.

## `enabled`

Must be `true` for the service to start.

## `port`

TCP port to listen on. Required — there is no default.

## `listen`

Bind address. Empty means "all interfaces" (`0.0.0.0` + `::`).

## `path`

HTTP path served. Defaults to `/`.

## `dump`

When `true`, every request (method, URI, headers) and response (status, headers, body) is logged at `debug` level. Off by default.

---

## Response formats

The endpoint accepts a `format` query parameter:

| Query | Response | Content-Type |
|---|---|---|
| (none) | Plain IP | `text/plain` |
| `?format=json` | `{"ip":"...","is_bogon":false}` | `application/json` |
| `?format=yaml` | `ip: ...\nis_bogon: false` | `application/yaml` |

`is_bogon` flags addresses that fall in private, reserved, or loopback ranges.