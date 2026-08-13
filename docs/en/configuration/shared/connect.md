This document was translated from Chinese by AI.

# ConnectOption

Network connection configuration.

```yaml
connect:
  fwmark: 0

  bindAddress4: ""
  bindAddress6: ""
  bindInterface: ""
  dialStrategy: prefer_ipv6
```

## `fwmark`

Linux `SO_MARK` firewall mark, used for policy routing and similar scenarios.

!!! warning "Supported platforms only"
    Linux

## `bindAddress4` / `bindAddress6`

Binds a local outgoing IP address. This specifies which IP sends requests on systems with multiple network interfaces.

```yaml
bindAddress4: "192.168.1.100"
bindAddress6: "::1"
```

## `bindInterface`

Binds to a specific network interface. The value can be an interface name such as `eth0` or an interface index such as `1`.

!!! warning "Supported platforms only"
    Linux, macOS, Windows

```yaml
# Bind by name
bindInterface: eth0

# Bind by index
bindInterface: 1
```

## `dialStrategy`

IP version (IPv4 / IPv6) selection strategy.

- **`prefer_ipv6`**: Prefer IPv6 while keeping IPv4 enabled (default)
- **`prefer_ipv4`**: Prefer IPv4 while keeping IPv6 enabled
- **`ipv6_only`**: Use IPv6 only
- **`ipv4_only`**: Use IPv4 only
