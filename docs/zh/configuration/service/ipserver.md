# IPServer

一个简单的 IP 地址 Echo 服务器。

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

??? note "行为说明"
    服务按下面的顺序读取请求 header，取第一个能解析为有效 IP 的值：`Cf-Connecting-IP`、`True-Client-IP`、`X-Real-IP`、`X-Forwarded-For`。若都没有，则回退到 TCP 远端地址。

    对于以逗号分隔的 header（典型如 `X-Forwarded-For`），取第一段。

## `enabled`

控制是否启用该服务。

## `port`

监听的 TCP 端口，可省略；省略时默认 `9002`。

## `listen`

监听地址。留空代表"所有接口"（`0.0.0.0` + `::`）。

## `path`

提供服务的 HTTP 路径。默认 `/`。

## `dump`

!!! note
    仅当全局日志 [`level`](../log.md) 为 `debug` 时，调试日志才会生效。


设为 `true` 时，每次请求（方法、URI、headers）和响应（状态码、headers、body）都会以 `debug` 级别打印日志。默认关闭。

---

## 响应格式

终结点接受 `format` 查询参数：

| Query          | 响应                            | Content-Type       |
|----------------|---------------------------------|--------------------|
| (不填)         | 纯文本 IP                       | `text/plain`       |
| `?format=json` | `{"ip":"...","is_bogon":false}` | `application/json` |
| `?format=yaml` | `ip: ...\nis_bogon: false`      | `application/yaml` |

`is_bogon` 用来标记地址是否落在私有、保留或环回范围内。
详见 [IPInfo Bogon](https://ipinfo.io/bogon)
