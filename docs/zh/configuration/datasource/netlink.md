# Netlink

从本地网络接口获取 IP 地址。适用于机器网卡已配置公网 IP 的场景。

```yaml
# required
type: netlink
name: data-netlink

# optional
ifName: eth0
ifIndex: 0
allowPrivate: false
```

??? note "行为说明"
    同时设置 `ifName` 和 `ifIndex` 时，优先使用 `ifIndex`；当索引查找失败时，回退到 `ifName`。返回匹配接口上的所有地址。

    该数据源同时返回 IPv4 和 IPv6 地址，可通过域名级别的 `ipv4` / `ipv6` 设置过滤。

---

## `ifName`

按网络接口名称过滤。

```yaml
ifName: eth0
```

## `ifIndex`

按网络接口索引过滤。若同时设置了 `ifName` 和 `ifIndex`，优先使用 `ifIndex`；当索引查找失败时，回退到 `ifName`。

```yaml
ifIndex: 2
```

## `allowPrivate`

是否允许返回私有 IP 地址（RFC 1918 等）。设为 `false`（默认）时，仅返回全球单播地址。

```yaml
allowPrivate: false
```
