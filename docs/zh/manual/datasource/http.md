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