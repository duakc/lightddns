# HTTP

通过 HTTP 请求获取 IP 地址。

```yaml
# required
type: http
name: data-http
url: https://api.ip.sb/ip

# optional
method: GET
headers:
  User-Agent: Lightddns/stable
match:
  jq: ""
  regex: ""

dns: system
connect:
  # ... ConnectOption
http:
  # ... HTTPOption
```

??? note "行为说明"
    HTTP 数据源会按 `connect.dialStrategy` 创建请求上下文：

    - `prefer_ipv6` / `prefer_ipv4`：同时创建 IPv4 和 IPv6 请求上下文，并合并结果。
    - `ipv4_only`：只创建 IPv4 请求上下文。
    - `ipv6_only`：只创建 IPv6 请求上下文。

    每个请求上下文使用相同的 `url`、`method`、`headers` 和 `match` 设置，但网络拨号器会固定为 `tcp4` 或 `tcp6`。

---

## `url`

获取 IP 地址的请求 URL。必须包含 `http://` 或 `https://` 协议头。

```yaml
url: https://api.ip.sb/ip
```

## `method`

HTTP 请求方法。留空时默认使用 `GET`。

规范值为 `GET`、`POST`、`PUT`、`HEAD`、`DELETE`、`PATCH`、`CONNECT`、
`OPTIONS`、`TRACE`、`BREW`、`PROPFIND`、`WHEN`。输入不区分大小写，加载
时会统一转换为大写。


## `headers`

自定义 HTTP 请求头。值可以是单个字符串或数组。如果没有配置 `User-Agent`，Lightddns 会添加默认 User-Agent。

```yaml
headers:
  User-Agent: Lightddns/stable
  Authorization: "Bearer xxxxx"
  X-Custom-Header:
    - value1
    - value2
```


## `match`

可选的 IP 提取规则。参见 [MatchOption](../shared/match.md)。

HTTP 数据源有一个特殊规则：当响应 `Content-Type` 是 JSON 且配置了 `match.jq` 时，不会再回退到正则或纯文本。这样可以避免 JSON 端点损坏后，被无关的响应文本误判为有效结果。


## `connect`

参见 [ConnectOption](../shared/connect.md)。

## `http`

参见 [HTTPOption](../shared/http.md)。

## `dns`

参见 [DNSOption](../shared/dns.md)。
