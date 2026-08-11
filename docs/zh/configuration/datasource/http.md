# HTTP

通过 HTTP 请求获取公网 IP 地址。支持纯文本响应、通过 `match.jq` 提取 JSON、通过 `match.regex` 提取文本，以及 IPv4/IPv6 连接选择。

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

## 常见服务

下面是几个常见公网 IP 服务的配置写法。服务可用性、频率限制、实际返回的地址族由这些外部服务自身决定。

### 纯文本

响应体只有 IP 地址时，可以不写 `match`。

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

JSON 端点使用 `match.jq`。

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
  url: https://api64.ipify.org?format=json
  match:
    jq: ".ip"
```

### 带说明文字的文本

页面里除了 IP 还有其它说明文字时，使用 `match.regex`。

```yaml
- type: http
  name: ipip
  url: https://myip.ipip.net
  match:
    regex: "当前 IP：\\s*(.+?)\\s*来自于："
```

---

## `url`

获取 IP 地址的请求 URL。必须包含 `http://` 或 `https://` 协议头。

```yaml
url: https://api.ip.sb/ip
```

---

## `method`

HTTP 请求方法。留空时默认使用 `GET`。

支持 `GET`、`POST`、`PUT`、`HEAD`、`DELETE`、`PATCH`、`CONNECT`、`OPTIONS`、`TRACE`，以及 `BREW`、`PROPFIND`、`WHEN`。

---

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

---

## `match`

可选的 IP 提取规则。参见 [MatchOption](../shared/match.md)。

HTTP 数据源有一个特殊规则：当响应 `Content-Type` 是 JSON 且配置了 `match.jq` 时，不会再回退到正则或纯文本。这样可以避免 JSON 端点损坏后，被无关的响应文本误判为有效结果。

---

## `connect`

参见 [ConnectOption](../shared/connect.md)。

## `http`

参见 [HTTPOption](../shared/http.md)。

## `dns`

参见 [DNSOption](../shared/dns.md)。
