This document was translated from Chinese by AI.

# Tencent Cloud

Updates DNS records through Tencent Cloud DNSPod.

```yaml
# required
type: tencentcloud
name: prov-tc
secretId: "{{ .Env.TENCENTCLOUD_SECRET_ID }}"
secretKey: "{{ .Env.TENCENTCLOUD_SECRET_KEY }}"

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
