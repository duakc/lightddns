# Cloudflare

A minimal Cloudflare setup: detect the public IPv4 via an HTTP service and keep an A record in sync.

```yaml
log:
  level: info

datasources:
  - type: http
    name: data-http
    url: https://api.ip.sb/ip
    dialStrategy: ipv4_only

providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-http
    interval: 30s
    timeout: 10s
    ttl: 60
    ipv4: true
```

Token permissions: `Zone:Read` + `DNS:Edit` on the zone you want to manage. Create one in the [Cloudflare token dashboard](https://dash.cloudflare.com/profile/api-tokens).

## Dual-stack

Drop `dialStrategy` and add an IPv6 endpoint to manage both A and AAAA records:

```yaml
datasources:
  - type: http
    name: data-http
    url:
      ipv4: https://api.ipify.org
      ipv6: https://api6.ipify.org

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-http
    ipv4: true
    ipv6: true
```

## Proxied records

Set `proxy: true` on the provider to enable Cloudflare's orange-cloud proxy on the records it manages:

```yaml
providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"
    proxy: true
```