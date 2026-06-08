# 阿里云

通过阿里云解析（alidns）更新 DNS 记录。

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

凭据需要 `AliyunDNSFullAccess` 策略（或一个仅包含 `DescribeDomainRecords`、`AddDomainRecord`、`UpdateDomainRecord`、`DeleteDomainRecord` 的自定义策略）。

!!! note
    阿里云解析会从 FQDN 提取父域 — `home.example.com` 对应的 Zone 是 `example.com`，请确认该 Zone 已经在阿里云解析控制台中存在。