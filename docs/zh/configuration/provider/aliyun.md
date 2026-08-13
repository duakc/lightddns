# 阿里云

通过阿里云解析（alidns）更新 DNS 记录。

手册: [Manual](../../manual/provider/aliyun.md)

```yaml
# required
type: aliyun
name: prov-ali
accessKeyId: "{{ .Env.ALIYUN_ACCESS_KEY_ID }}"
accessKeySecret: "{{ .Env.ALIYUN_ACCESS_KEY_SECRET }}"

# optional
dns: system
connect:
  # ... ConnectOption
http:
  # ... HTTPOption
```

## `accessKeyId`

阿里云 RAM 访问密钥 ID。

## `accessKeySecret`

阿里云 RAM 访问密钥 Secret。

!!! note "权限要求"
    凭据需要对目标 Zone 的 `DescribeDomainRecords`、`AddDomainRecord`、`UpdateDomainRecord`、`DeleteDomainRecord` 调用权限。预置策略 `AliyunDNSFullAccess` 即可满足。

!!! note "父域识别"
    阿里云解析会从 FQDN 提取父域 — `home.example.com` 对应的 Zone 是 `example.com`，请确认该 Zone 已经在阿里云解析控制台中存在。

---

## `connect`

参见 [ConnectOption](../shared/connect.md)。

## `http`

参见 [HTTPOption](../shared/http.md)。

## `dns`

参见 [DNSOption](../shared/dns.md)。