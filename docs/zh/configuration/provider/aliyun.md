# 阿里云

通过阿里云解析（alidns）更新 DNS 记录。

手册: [Manual](../../manual/provider/aliyun.md)

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

阿里云 RAM 访问密钥 ID。

## `accessKeySecret`

阿里云 RAM 访问密钥 Secret。

## `lines`

要为 DNS 记录指定线路时使用字符串或字符串数组。未配置或使用空数组时使用 `default`；只有一个值时，该线路用于所有 IP；有多个值时，每条线路都会应用到每个 IP。

只配置一条非 `default` 线路是允许的，但前提是所需的 `default` 记录已经存在。阿里云要求同一个主机名和记录类型先存在 `default` 线路记录，才能创建或更新其它线路。空 Zone 初始化时，请先创建 `default` 记录（例如先只配置 `default`），再加入其它线路；同时建议把 `default` 保留在线路列表中，使后续启动时也会继续更新它。若 API 返回线路错误，Lightddns 会保留云端错误并提示先创建 `default` 线路记录。

线路代码请以阿里云 API 文档为准，参见 [AddDomainRecord 的 Line 参数](https://help.aliyun.com/document_detail/29772.html)；控制台显示名称不一定等于 API 代码。

生成的配置 schema 会提供常用线路代码用于编辑器补全；阿里云支持的其它线路值也仍然有效。

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
