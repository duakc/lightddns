This document was translated from Chinese by AI.

# Aliyun

Updates DNS records through Aliyun DNS (alidns).

Manual: [Manual](../../manual/provider/aliyun.md)

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

Aliyun RAM access key ID.

## `accessKeySecret`

Aliyun RAM access key secret.

!!! note "Required permissions"
    The credentials need permission to call `DescribeDomainRecords`, `AddDomainRecord`, `UpdateDomainRecord`, and `DeleteDomainRecord` for the target zone. The predefined `AliyunDNSFullAccess` policy provides these permissions.

!!! note "Parent domain detection"
    Aliyun DNS extracts the parent domain from the FQDN: the zone for `home.example.com` is `example.com`. Make sure this zone already exists in the Aliyun DNS console.

---

## `connect`

See [ConnectOption](../shared/connect.md).

## `http`

See [HTTPOption](../shared/http.md).

## `dns`

See [DNSOption](../shared/dns.md).
