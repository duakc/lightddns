# Aliyun

Updates DNS records via Aliyun DNS (alidns).

```yaml
log:
  level: info

datasources:
  - type: http
    name: data-http
    url: https://api.ip.sb/ip
    dialStrategy: ipv4_only

providers:
  - type: aliyun
    name: prov-ali
    accessKeyId: "{{ .Env.ALIYUN_ACCESS_KEY_ID }}"
    accessKeySecret: "{{ .Env.ALIYUN_ACCESS_KEY_SECRET }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-ali
    datasource: data-http
    interval: 30s
    timeout: 10s
    ttl: 600
    ipv4: true
```

The credentials need permissions on the `AliyunDNSFullAccess` policy (or a narrower custom policy that allows `DescribeDomainRecords`, `AddDomainRecord`, `UpdateDomainRecord`, and `DeleteDomainRecord` on the parent zone).

!!! note
    Aliyun DNS resolves the parent zone from the FQDN — `home.example.com` maps to the zone `example.com`. The zone must already exist in your Aliyun DNS console.