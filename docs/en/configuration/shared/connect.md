# ConnectOption

Network connection configuration, shared by datasources and Service Providers. These fields live under the `connect:` key of a datasource or provider.

```yaml
connect:
  fwmark: 0

  bindAddress4: ""
  bindAddress6: ""
  bindInterface: ""
  dialStrategy: prefer_ipv6
```

## `fwmark`

Linux SO_MARK firewall mark for policy routing.

!!! warning "Only available on below platform"
    Linux

## `bindAddress4` / `bindAddress6`

Bind to a local outgoing IP address. Useful in multi-homed scenarios.

```yaml
bindAddress4: "192.168.1.100"
bindAddress6: "::1"
```

## `bindInterface`

Bind to a specific network interface. Accepts an interface name (e.g. `eth0`) or interface index (e.g. `1`).

!!! warning "Only available on below platform"
    Linux, Macos, Windows

```yaml
# By name
bindInterface: eth0

# By index
bindInterface: 1
```

## `dialStrategy`

IP version (IPv4 / IPv6) selection strategy.

- **`prefer_ipv6`** — Prefer IPv6, IPv4 enabled (default)
- **`prefer_ipv4`** — Prefer IPv4, IPv6 enabled
- **`ipv6_only`** — IPv6 only
- **`ipv4_only`** — IPv4 only
