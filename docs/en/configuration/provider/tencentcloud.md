This document was translated from Chinese by AI.

# Tencent Cloud

Updates DNS records through Tencent Cloud DNSPod.

```yaml
# required
type: tencentcloud
name: prov-tc
secretId: "{{ .Env.TENCENTCLOUD_SECRET_ID }}"
secretKey: "{{ .Env.TENCENTCLOUD_SECRET_KEY }}"
# optional: one line, or multiple lines applied to every address
lines: [默认]

# optional
dns: system
connect:
  # ... ConnectOption
http:
  # ... HTTPOption
```

## `secretId`

Tencent Cloud CAM Secret ID.

## `secretKey`

Tencent Cloud CAM Secret Key.

## `lines`

Use a string or string array to assign DNS record lines. An omitted value or empty array uses `默认`; one value is shared by all IP addresses; multiple values create one record for every configured line and every IP address.

A single non-default line is valid after the required `默认` record already exists. Tencent Cloud requires a `默认` line record for the same hostname and record type before other line records can be created or updated. On an empty zone, create the `默认` record first (for example, initialize with only `默认`), then add other lines; keep `默认` in the configured lines so it continues to be updated on later startups. If the API returns a line error, Lightddns preserves the provider error and adds a message telling you to create the `默认` line record first.

Use the Tencent Cloud API documentation as the source of truth for line codes: see the [`RecordLine` parameter in CreateRecord](https://cloud.tencent.com/document/api/1427/56180). Console display names are not always the API values.

The generated configuration schema offers common line codes for editor completion; other values supported by Tencent Cloud remain valid.

!!! note "Required permissions"
    The credentials need permission to call `DescribeRecordList`, `CreateRecord`, `ModifyRecord`, and `DeleteRecord`. The predefined `QcloudDNSPodFullAccess` policy provides these permissions.

!!! note "Empty zone"
    A new zone without any records returns `NoDataOfRecord`. Lightddns treats this as an empty record set and automatically creates the first record in the next step.

---

## `connect`

See [ConnectOption](../shared/connect.md).

## `http`

See [HTTPOption](../shared/http.md).

## `dns`

See [DNSOption](../shared/dns.md).
