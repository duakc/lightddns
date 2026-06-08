# 腾讯云

通过腾讯云 DNSPod 更新 DNS 记录。

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

凭据需要 `DescribeRecordList`、`CreateRecord`、`ModifyRecord`、`DeleteRecord` 的调用权限，预置策略 `QcloudDNSPodFullAccess` 即可。

!!! note
    对全新、还没有任何记录的 Zone，腾讯云 DNSPod 会返回 `NoDataOfRecord`。Lightddns 把它当作空记录集，下一步会自动创建记录，无需手动预先添加一条占位记录。