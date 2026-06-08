# Tencent Cloud

Updates DNS records via Tencent Cloud DNSPod.

```yaml
log:
  level: info

datasources:
  - type: http
    name: data-http
    url: https://api.ip.sb/ip
    dialStrategy: ipv4_only

providers:
  - type: tencentcloud
    name: prov-tc
    secretId: "{{ .Env.TENCENTCLOUD_SECRET_ID }}"
    secretKey: "{{ .Env.TENCENTCLOUD_SECRET_KEY }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-tc
    datasource: data-http
    interval: 30s
    timeout: 10s
    ttl: 600
    ipv4: true
```

The credentials need permission to call `DescribeRecordList`, `CreateRecord`, `ModifyRecord`, and `DeleteRecord` on the parent zone. The pre-defined policy `QcloudDNSPodFullAccess` is sufficient.

!!! note
    Tencent Cloud DNSPod returns `NoDataOfRecord` on a brand-new zone with no records. Lightddns treats this as an empty record set and proceeds to create the record — no manual seeding required.