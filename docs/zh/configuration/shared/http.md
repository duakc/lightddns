# HTTPOption

HTTP 客户端配置。

```yaml
http:
  useSystemProxy: false

  httpProxy: ""
  httpsProxy: ""

  httpDebug: false
```

## `useSystemProxy`

是否使用系统代理。若配置了 `httpProxy` 或 `httpsProxy`，该选项会被忽略。

## `httpProxy` / `httpsProxy`

分别指定 HTTP 和 HTTPS 代理地址。若只填写其中一个，另一个会自动使用相同的值。

```yaml
httpProxy: "http://127.0.0.1:7890"
httpsProxy: "http://127.0.0.1:7890"
```

## `httpDebug`

设为 `true` 时，会记录完整的 HTTP 请求与响应（方法、URL、请求头、请求体）。默认关闭。

!!! note
    仅当全局日志 [`level`](../log.md) 为 `debug` 时，调试日志才会生效。

!!! warning "敏感信息与性能"
    HTTP 调试日志会暴露敏感信息 —— API 令牌、凭据以及请求/响应体都会被原样写入日志。
    同时它会导致**严重的性能下降**，因为每个请求和响应都会被完整缓冲并序列化。
    非必要请勿开启；确需调试时再打开，并且在把日志发送给他人之前，务必先检查其中是否包含隐私数据。

```yaml
httpDebug: true
```
