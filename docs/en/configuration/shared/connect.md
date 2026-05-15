# ConnectOption

Network connection configuration, shared by datasources and Service Providers.

```yaml
fwmark: 0
dns: system

bindAddress4: ""
bindAddress6: ""
bindInterface: ""
dialStrategy: prefer_ipv6
```

## `fwmark`

Linux SO_MARK firewall mark for policy routing.

!!! warning "Only available on below platform"
    Linux

## `dns`

DNS resolution configuration.

`system` uses the system DNS resolver.

```yaml
dns: system
```

Customize the upstream DNS server via DNS over TLS. `type` accepts `system | tls`.

```yaml
dns:
  type: tls
  server: 8.8.8.8
  port: 853
```

!!! note
    DNS config only takes effect when the target is a domain name. It is skipped when the target is an IP address.

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

IP address family selection strategy.

- **`prefer_ipv6`** — Prefer IPv6, IPv4 enabled (default)
- **`prefer_ipv4`** — Prefer IPv4, IPv6 enabled
- **`ipv6_only`** — IPv6 only
- **`ipv4_only`** — IPv4 only
