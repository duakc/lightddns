# Netlink

Retrieves IP addresses from local network interfaces. Useful when the machine already has a public IP assigned to one of its interfaces.

```yaml
# required
type: netlink
name: data-netlink

# optional
ifName: eth0
ifIndex: 0
allowPrivate: false
```

??? note "Behavior"
    When both `ifName` and `ifIndex` are set, `ifIndex` takes priority; `ifName` is used as a fallback if the index lookup fails. All addresses on the matching interface are returned.

    This datasource returns both IPv4 and IPv6 addresses. Use the domain-level `ipv4` / `ipv6` settings to filter.

---

## `ifName`

Filter by network interface name.

```yaml
ifName: eth0
```

## `ifIndex`

Filter by network interface index. When both `ifName` and `ifIndex` are set, `ifIndex` takes priority; `ifName` is used as a fallback if the index lookup fails.

```yaml
ifIndex: 2
```

## `allowPrivate`

Whether to include private IP addresses (RFC 1918, etc.). When `false` (default), only global unicast addresses are returned.

```yaml
allowPrivate: false
```
