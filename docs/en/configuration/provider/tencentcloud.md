# Tencent Cloud

Updates DNS records via Tencent Cloud DNSPod.

```yaml
# required
type: tencentcloud
name: prov-tc
secretId: "{{ .Env.TENCENTCLOUD_SECRET_ID }}"
secretKey: "{{ .Env.TENCENTCLOUD_SECRET_KEY }}"

# ... ConnectOption
# ... HTTPOption
```

## `secretId`

Tencent Cloud CAM secret ID.

## `secretKey`

Tencent Cloud CAM secret key.

!!! note "Permissions"
    The credentials need `DescribeRecordList`, `CreateRecord`, `ModifyRecord`, and `DeleteRecord` on the parent zone. The pre-defined policy `QcloudDNSPodFullAccess` is sufficient.

!!! note "Empty zone"
    A brand-new zone with no records returns `NoDataOfRecord`. Lightddns treats this as an empty record set and proceeds to create the first record automatically.

---

## `ConnectOption`

See [ConnectOption](../shared/connect.md).

## `HTTPOption`

See [HTTPOption](../shared/http.md).