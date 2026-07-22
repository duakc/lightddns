# Aliyun

Updates DNS records via Aliyun DNS (alidns).

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

!!! note "Permissions"
    The credentials need `DescribeDomainRecords`, `AddDomainRecord`, `UpdateDomainRecord`, and `DeleteDomainRecord` on the parent zone. The pre-defined policy `AliyunDNSFullAccess` is sufficient.

!!! note "Zone resolution"
    Aliyun DNS resolves the parent zone from the FQDN — `home.example.com` maps to the zone `example.com`. The zone must already exist in your Aliyun DNS console.

---

## `connect`

See [ConnectOption](../shared/connect.md).

## `http`

See [HTTPOption](../shared/http.md).

## `dns`

See [DNSOption](../shared/dns.md).