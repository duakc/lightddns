# Group datasources

`sum` and `failover` are datasources that take other datasources as input. They let you combine multiple discovery methods without changing how domains reference them.

## Sum

Collects IPs from every child and merges the results. Useful when IPv4 and IPv6 come from different endpoints.

```yaml
datasources:
  - type: http
    name: data-v4
    url: https://api.ipify.org
    dialStrategy: ipv4_only

  - type: http
    name: data-v6
    url: https://api6.ipify.org
    dialStrategy: ipv6_only

  - type: sum
    name: data-all
    datasources:
      - data-v4
      - data-v6

providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-all
    ipv4: true
    ipv6: true
```

By default `sum` only errors when **every** child fails. Set `fastFail: true` to short-circuit on the first error.

## Failover

Tries children in order until one succeeds. Use it to fall back from a fast local source to a slower remote one.

```yaml
datasources:
  - type: netlink
    name: data-nic
    ifName: eth0

  - type: http
    name: data-http
    url: https://api.ip.sb/ip

  - type: failover
    name: data-fo
    datasources:
      - data-nic    # try the interface first
      - data-http   # fall back to the public API

providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-fo
```

Once a child succeeds, the next cycle starts from that same child — there's no penalty for keeping a long list of fallbacks.