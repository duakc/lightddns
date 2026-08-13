# ConnectOption

网络连接配置。

```yaml
connect:
  fwmark: 0

  bindAddress4: ""
  bindAddress6: ""
  bindInterface: ""
  dialStrategy: prefer_ipv6
```

## `fwmark`

Linux SO_MARK 防火墙标记，用于策略路由等场景。

!!! warning "仅支持以下平台"
    Linux

## `bindAddress4` / `bindAddress6`

绑定本地出口 IP 地址。在多网卡场景下指定从哪个 IP 发出请求。

```yaml
bindAddress4: "192.168.1.100"
bindAddress6: "::1"
```

## `bindInterface`

绑定到指定的网络接口。可以是接口名（如 `eth0`）或接口索引（如 `1`）。

!!! warning "仅支持以下平台"
    Linux, Macos, Windows

```yaml
# 按名称绑定
bindInterface: eth0

# 按索引绑定
bindInterface: 1
```

## `dialStrategy`

IP 版本（IPv4 / IPv6）选择策略。

- **`prefer_ipv6`** — 优先使用 IPv6，同时启用 IPv4（默认）
- **`prefer_ipv4`** — 优先使用 IPv4，同时启用 IPv6
- **`ipv6_only`** — 仅使用 IPv6
- **`ipv4_only`** — 仅使用 IPv4
