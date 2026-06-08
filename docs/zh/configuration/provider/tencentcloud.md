# 腾讯云

通过腾讯云 DNSPod 更新 DNS 记录。

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

腾讯云 CAM Secret ID。

## `secretKey`

腾讯云 CAM Secret Key。

!!! note "权限要求"
    凭据需要 `DescribeRecordList`、`CreateRecord`、`ModifyRecord`、`DeleteRecord` 的调用权限。预置策略 `QcloudDNSPodFullAccess` 即可满足。

!!! note "空 Zone"
    全新、还没有任何记录的 Zone 会返回 `NoDataOfRecord`。Lightddns 把它当作空记录集处理，下一步会自动创建第一条记录。

---

## `ConnectOption`

参见 [ConnectOption](../shared/connect.md)。

## `HTTPOption`

参见 [HTTPOption](../shared/http.md)。