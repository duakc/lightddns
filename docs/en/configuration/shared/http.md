# HTTPOption

HTTP client configuration, shared by datasources and Service Providers. These fields live under the `http:` key of a datasource or provider.

```yaml
http:
  useSystemProxy: false

  httpProxy: ""
  httpsProxy: ""

  httpDebug: false
```

## `useSystemProxy`

Whether to use the system proxy. Ignored when `httpProxy` or `httpsProxy` is set.

## `httpProxy` / `httpsProxy`

HTTP and HTTPS proxy addresses respectively. If only one is set, the other defaults to the same value.

```yaml
httpProxy: "http://127.0.0.1:7890"
httpsProxy: "http://127.0.0.1:7890"
```

## `httpDebug`

When `true`, the full HTTP request and response (method, URL, headers, body) are logged. Off by default.

!!! note
    Debug logging only takes effect when the global log [`level`](../log.md) is `debug`.

!!! warning "Sensitive data & performance"
    HTTP debug logging exposes sensitive information — API tokens, credentials, and request/response bodies are written to the log verbatim. It also causes a **severe performance drop** because every request and response is fully buffered and serialized. Leave it off unless you are actively debugging, and when sharing the resulting logs with others, review them for private data first.

```yaml
httpDebug: true
```
