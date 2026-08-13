This document was translated from Chinese by AI.

# IPServer

A simple IP address echo server.

```yaml
# required
type: ipserver
name: svc-ip
enabled: true

# optional
listen: ""
port: 9002
path: /
dump: false
```

??? note "Behavior"
    The service reads request headers in the following order and uses the first value that can be parsed as a valid IP address: `Cf-Connecting-IP`, `True-Client-IP`, `X-Real-IP`, `X-Forwarded-For`. If none contains a valid address, it falls back to the remote TCP address.

    For comma-separated headers, such as `X-Forwarded-For`, it uses the first segment.

## `enabled`

Controls whether the service is enabled.

## `port`

The TCP listening port. Optional; defaults to `9002` when omitted.

## `listen`

The listening address. An empty value means all interfaces (`0.0.0.0` and `::`).

## `path`

The HTTP path served by the service. Defaults to `/`.

## `dump`

!!! note
    Debug logs only take effect when the global log [`level`](../log.md) is `debug`.

When set to `true`, every request (method, URI, headers) and response (status code, headers, body) is logged at the `debug` level. Disabled by default.

---

## Response Format

The endpoint accepts a `format` query parameter:

| Query          | Response                        | Content-Type       |
|----------------|---------------------------------|--------------------|
| (omitted)      | Plain-text IP                   | `text/plain`       |
| `?format=json` | `{"ip":"...","is_bogon":false}` | `application/json` |
| `?format=yaml` | `ip: ...\nis_bogon: false`      | `application/yaml` |

`is_bogon` indicates whether the address belongs to a private, reserved, or loopback range.
See [IPInfo Bogon](https://ipinfo.io/bogon) for details.
