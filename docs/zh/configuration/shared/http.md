# HTTPOption

HTTP 客户端配置，被数据源和服务提供者复用。

```yaml
useSystemProxy: false

httpProxy: ""
httpsProxy: ""
```

## `useSystemProxy`

是否使用系统代理。若配置了 `httpProxy` 或 `httpsProxy`，此选项会被忽略。

## `httpProxy` / `httpsProxy`

分别指定 HTTP 和 HTTPS 代理地址。若只填写其中一个，另一个会自动使用相同的值。

```yaml
httpProxy: "http://127.0.0.1:7890"
httpsProxy: "http://127.0.0.1:7890"
```
