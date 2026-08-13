This document was translated from Chinese by AI.

# HTTPOption

HTTP client configuration.

```yaml
http:
  useSystemProxy: false

  httpProxy: ""
  httpsProxy: ""

  httpDebug: false
```

## `useSystemProxy`

Whether to use the system proxy. This option is ignored if `httpProxy` or `httpsProxy` is configured.

## `httpProxy` / `httpsProxy`

The HTTP and HTTPS proxy addresses, respectively. If only one is specified, the other automatically uses the same value.

```yaml
httpProxy: "http://127.0.0.1:7890"
httpsProxy: "http://127.0.0.1:7890"
```

## `httpDebug`

When set to `true`, complete HTTP requests and responses (method, URL, request headers, request body) are logged. Disabled by default.

!!! note
    Debug logs only take effect when the global log [`level`](../log.md) is `debug`.

!!! warning "Sensitive information and performance"
    HTTP debug logs expose sensitive information: API tokens, credentials, and request and response bodies are written to the log as-is.
    They also cause a **severe performance degradation**, because every request and response is fully buffered and serialized.
    Do not enable this unless necessary. Enable it only while debugging, and always check logs for private information before sharing them.

```yaml
httpDebug: true
```
