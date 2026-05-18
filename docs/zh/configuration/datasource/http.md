# HTTP

通过 HTTP 请求获取本机公网 IP 地址。支持 JSON（jq）解析、正则提取，以及双栈网络（IPv4/IPv6）。

```yaml
# required
type: http
name: data-http
url: https://api64.ipify.org?format=json

# optional
json:
  ipv4: ".ip"
  ipv6: ".ip"
regex:
  ipv4: ""
  ipv6: ""
method: GET
headers:
  User-Agent: Lightddns/stable

# ... ConnectOption
# ... HTTPOption
```

??? note "行为说明"
    HTTP 数据源根据 `dialStrategy` 分别创建 IPv4 和 IPv6 的请求上下文：

    - 非 `ipv4_only` 模式创建 IPv6 上下文（`tcp6` 拨号）
    - 非 `ipv6_only` 模式创建 IPv4 上下文（`tcp4` 拨号）

    各自的 `json`/`regex` 表达式独立提取对应 IP，最终合并返回。

---

[JQ表达式和Regex部分示例](../example/http.md)

## `url`

获取 IP 地址的请求 URL。**必须携带 `http://` 或 `https://` 协议头**。

```yaml
# 返回纯文本 IP
url: https://api.ip.sb/ip

# 返回 JSON
url: https://api64.ipify.org?format=json
```

若 URL 主机部分是 IP 地址而非域名，`dialStrategy` 会自动覆盖为该 IP 对应的版本（IPv4 → `ipv4_only`，IPv6 → `ipv6_only`）。

---

## `json`

使用 [jq](https://github.com/itchyny/gojq) 语法从 JSON 响应中提取 IP 地址。

**简写形式**（`string`）— IPv4 和 IPv6 共用同一个 jq 表达式：

```yaml
json: ".ip"
```

**对象形式** — 单独为 IPv4 和 IPv6 指定不同的 jq 表达式：

```yaml
json:
  ipv4: ".ipv4"
  ipv6: ".ipv6"
```

当响应 `Content-Type` 为 `application/json` 且配置了 `json` 字段时，优先使用 JSON 路径提取。jq 查询结果会被解析为 IP 地址，请确保返回值是合法的 IPv4 或 IPv6 字符串。

---

## `regex`

使用正则表达式从非 JSON 响应中提取 IP 地址。与 `json` 至少配置一个。

**简写形式**（`string`）— IPv4 和 IPv6 共用同一个正则：

```yaml
regex: "IP:\\s*(.+?)\\n"
```

**对象形式** — 单独为 IPv4 和 IPv6 指定不同的正则：

```yaml
regex:
  ipv4: "IPv4:\\s*(.+?)\\n"
  ipv6: "IPv6:\\s*(.+?)\\n"
```

取**匹配到的第一个**作为 IP 地址。

!!! note "提取优先级"
    1. 若响应 `Content-Type` 为 `application/json` 且配置了 `json` → JSON（jq）提取
    2. 若配置了 `regex` → 正则提取
    3. 否则 → 将响应体整体作为纯文本 IP 解析

---

## `method`

HTTP 请求方法。支持 `GET`、`POST`、`PUT`、`HEAD`、`DELETE`、`PATCH`、`CONNECT`、`OPTIONS`、`TRACE`，以及 `BREW`、`PROPFIND`、`WHEN`。

---

## `headers`

自定义 HTTP 请求头。值可以是单个字符串或数组。

```yaml
headers:
  User-Agent: Lightddns/stable
  Authorization: "Bearer xxxxx"
  X-Custom-Header:
    - value1
    - value2
```

---

## `ConnectOption`

参见 [ConnectOption](../shared/connect.md)。

## `HTTPOption`

参见 [HTTPOption](../shared/http.md)。
