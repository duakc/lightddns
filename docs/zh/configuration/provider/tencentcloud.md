# 腾讯云

通过腾讯云 DNSPod 更新 DNS 记录。

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

腾讯云 CAM Secret ID。

## `secretKey`

腾讯云 CAM Secret Key。

## `lines`

要为 DNS 记录指定线路时使用字符串或字符串数组。未配置或使用空数组时使用 `默认`；只有一个值时，该线路用于所有 IP；有多个值时，每条线路都会应用到每个 IP。

只配置一条非 `默认` 线路是允许的，但前提是所需的 `默认` 记录已经存在。腾讯云要求同一个主机名和记录类型先存在 `默认` 线路记录，才能创建或更新其它线路。空 Zone 初始化时，请先创建 `默认` 记录（例如先只配置 `默认`），再加入其它线路；同时建议把 `默认` 保留在线路列表中，使后续启动时也会继续更新它。若 API 返回线路错误，Lightddns 会保留云端错误并提示先创建 `默认` 线路记录。

线路代码请以腾讯云 API 文档为准，参见 [CreateRecord 的 RecordLine 参数](https://cloud.tencent.com/document/api/1427/56180)；控制台显示名称不一定等于 API 代码。

生成的配置 schema 会提供常用线路代码用于编辑器补全；腾讯云支持的其它线路值也仍然有效。

!!! note "权限要求"
    凭据需要 `DescribeRecordList`、`CreateRecord`、`ModifyRecord`、`DeleteRecord` 的调用权限。预置策略 `QcloudDNSPodFullAccess` 即可满足。

!!! note "空 Zone"
    全新、还没有任何记录的 Zone 会返回 `NoDataOfRecord`。Lightddns 把它当作空记录集处理，下一步会自动创建第一条记录。

---

## `connect`

参见 [ConnectOption](../shared/connect.md)。

## `http`

参见 [HTTPOption](../shared/http.md)。

## `dns`

参见 [DNSOption](../shared/dns.md)。
