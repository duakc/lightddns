This document was translated from Chinese by AI.

# Netlink

Obtains IP addresses from local network interfaces. This is suitable when a public IP address is already configured on a network interface of the machine.

```yaml
# required
type: netlink
name: data-netlink

# optional
ifName: eth0
ifIndex: 0
includeBogon: false
```

??? note "Behavior"
    When both `ifName` and `ifIndex` are set, `ifIndex` takes priority. If the index lookup fails, it falls back to `ifName`. All addresses on the matching interface are returned.

    This datasource returns both IPv4 and IPv6 addresses. Use the domain-level `ipv4` / `ipv6` settings to filter them.

---

## `ifName`

Filters by network interface name. At least one of `ifName` and `ifIndex` must be set.

```yaml
ifName: eth0
```

## `ifIndex`

Filters by network interface index. At least one of `ifName` and `ifIndex` must be set. When both are configured, `ifIndex` takes priority; if the index lookup fails, it falls back to `ifName`.

```yaml
ifIndex: 2
```

## `includeBogon`

Whether bogon addresses may be returned, including private (RFC 1918), link-local, loopback, and other reserved ranges. When set to `false` (the default), only global unicast addresses are returned.
See [IPInfo Bogon](https://ipinfo.io/bogon) for details.

```yaml
includeBogon: false
```
