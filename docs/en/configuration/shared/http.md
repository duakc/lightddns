# HTTPOption

HTTP client configuration, shared by datasources and Service Providers.

```yaml
useSystemProxy: false

httpProxy: ""
httpsProxy: ""
```

## `useSystemProxy`

Whether to use the system proxy. Ignored when `httpProxy` or `httpsProxy` is set.

## `httpProxy` / `httpsProxy`

HTTP and HTTPS proxy addresses respectively. If only one is set, the other defaults to the same value.

```yaml
httpProxy: "http://127.0.0.1:7890"
httpsProxy: "http://127.0.0.1:7890"
```
