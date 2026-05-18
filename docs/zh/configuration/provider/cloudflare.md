# Cloudflare

通过 [Cloudflare API](https://developers.cloudflare.com/api/) 更新 DNS 记录。
可选开启代理（橙色云）模式。

```yaml
# required
type: cloudflare
name: prov-cf
token: "your-cloudflare-api-token"

# optional
proxy: false

# ... ConnectOption
# ... HTTPOption
```

## `token`

Cloudflare API 令牌。在 [Cloudflare 控制台](https://dash.cloudflare.com/profile/api-tokens) 创建，需要以下权限：

- **Zone** — `DNS:Edit`
- **Zone** — `Zone:Read`

```yaml
token: "your-cloudflare-api-token"
```

---

## `proxy`

是否开启 Cloudflare 代理（橙色云）。

```yaml
proxy: true
```

---

## `ConnectOption`

参见 [ConnectOption](../shared/connect.md)。

## `HTTPOption`

参见 [HTTPOption](../shared/http.md)。
