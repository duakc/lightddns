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
# optional: one line, or multiple lines applied to every address
lines: [default]

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

## `lines`

Use a string or string array to assign DNS record lines. An omitted value or empty array uses `default`; one value is shared by all IP addresses; multiple values create one record for every configured line and every IP address.

A single non-default line is valid after the required `default` record already exists. Aliyun requires a `default` line record for the same hostname and record type before other line records can be created or updated. On an empty zone, create the `default` record first (for example, initialize with only `default`), then add other lines; keep `default` in the configured lines so it continues to be updated on later startups. If the API returns a line error, Lightddns preserves the provider error and adds a message telling you to create the `default` line record first.

Use the Aliyun API documentation as the source of truth for line codes: see the [`Line` parameter in AddDomainRecord](https://help.aliyun.com/document_detail/29772.html). Console display names are not always the API values.

The generated configuration schema offers common line codes for editor completion; other values supported by Aliyun remain valid.

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
