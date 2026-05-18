# 快速开始

本指南带你完成 Lightddns 的首次配置，让家里的域名能自动更新 IP。整个过程大约 5 分钟。

## 准备工作

- 一个由 **Cloudflare** 管理的域名（其他服务商即将支持）
- 一个具有 DNS 编辑权限的 Cloudflare **API 令牌**
- 已安装 Lightddns（参见[安装指南](installation.md)）

### 获取 Cloudflare API 令牌

1. 登录 [Cloudflare 控制台](https://dash.cloudflare.com)
2. 进入 **My Profile** → **API Tokens**（或[点此直达](https://dash.cloudflare.com/profile/api-tokens)）
3. 点击 **Create Token**
4. 使用 **Edit zone DNS** 模板，或创建自定义令牌，权限如下：
    - `Zone` — `DNS` — `Edit`
    - `Zone` — `Zone` — `Read`
5. 在 **Zone Resources** 中选择要管理的域名
6. 点击 **Continue to summary** → **Create Token**
7. 复制令牌——稍后需要填入配置文件

## 第一步：创建配置文件

新建一个配置文件。建议放在 `/etc/lightddns/lightddns.yaml`，但任意路径均可。

```yaml
# lightddns.yaml

datasources:
  # 通过免费的 HTTP API 获取公网 IPv4 地址
  - type: http
    name: my-public-ip
    url: https://api64.ipify.org?format=json
    json: ".ip"

providers:
  # 连接 Cloudflare 以更新 DNS 记录
  - type: cloudflare
    name: my-cloudflare
    token: "{{.Env.CF_API_TOKEN}}"

domains:
  # 将上述两者绑定：检查 IP，更新 home.yourdomain.com
  - domain: home.yourdomain.com
    provider: my-cloudflare
    datasource: my-public-ip

    # 可选：自定义更新行为
    interval: 5m       # 每 5 分钟检查一次
    ttl: 300           # DNS TTL 5 分钟
    ipv4: true         # 更新 A 记录
    ipv6: false        # 不更新 AAAA 记录
```

将 `home.yourdomain.com` 替换为你实际的域名。

!!! warning "域名须已在 Cloudflare 中存在"
    Lightddns 只能更新已有的 DNS 记录。在运行之前，先登录 Cloudflare 为 `home.yourdomain.com` 创建一条 A 记录，IP 随意填写（如 `1.2.3.4`）。第一次运行时，Lightddns 会将其更新为你的真实 IP。

## 第二步：设置 API 令牌

在配置文件旁边创建 `.env` 文件（或直接导出环境变量）：

```bash
# .env
CF_API_TOKEN=你的-cloudflare-api-令牌
```

配置文件中的 `{{.Env.CF_API_TOKEN}}` 会被自动替换为环境变量或 `.env` 文件中的值。这样密钥就不会暴露在配置文件中了。

## 第三步：测试运行

先执行一次，验证配置是否正确：

```bash
lightddns run -c lightddns.yaml --once
```

如果看到类似这样的日志输出，说明一切正常：

```
INFO  main  ipv4 ip updated  domain=home.yourdomain.com  ip=203.0.113.5
```

在 Cloudflare 控制台中查看，DNS 记录应该已经更新为你的实际公网 IP。

## 第四步：长期运行

去掉 `--once` 参数，让程序持续运行：

```bash
lightddns run -c lightddns.yaml
```

Lightddns 会按设定的间隔检查 IP，仅在变化时更新 DNS。按 `Ctrl+C` 停止。

正式使用时，建议配置为 [systemd 服务](installation.md#_2)以实现开机自启和后台运行。

## 自定义 IP 获取方式

上面的示例使用了 [ipify.org](https://www.ipify.org)——一个免费可靠的 IP 查询服务。你也可以使用其他服务。

### 纯文本 API

有些服务直接返回纯文本 IP：

```yaml
datasources:
  - type: http
    name: ipv4-http
    url: https://api.ip.sb/ip
    # 无需配置 json/regex——响应正文直接作为 IP 使用
```

### 双栈（IPv4 + IPv6）

同时获取 IPv4 和 IPv6 地址时，分别指定提取规则：

```yaml
datasources:
  - type: http
    name: dualstack-http
    url: https://api64.ipify.org?format=json
    json:
      ipv4: ".ip"
    # IPv6 需要使用返回你 IPv6 地址的服务。
    # 多数服务每次请求只返回一个 IP 版本。
```

??? info "双栈工作原理"
    Lightddns 会为 HTTP 数据源分别建立 IPv4 和 IPv6 连接。如果你的网络同时支持两种协议，就能获取到两个地址。详见 [ConnectOption 参考](../configuration/shared/connect.md)中的 `dialStrategy` 选项。

### 使用脚本

如果想通过自定义脚本获取 IP：

```yaml
datasources:
  - type: command
    name: my-script
    cmd:
      ipv4: "curl -s https://api.ip.sb/ip"
```

命令的输出会被自动扫描提取有效的 IP 地址。

### 读取网卡地址

如果你的机器网卡上直接绑定了公网 IP：

```yaml
datasources:
  - type: netlink
    name: eth0-ip
    ifName: eth0
    allowPrivate: false   # 仅返回公网 IP
```

## 管理多个域名

你可以管理任意数量的域名——只需在 `domains` 下添加更多条目：

```yaml
domains:
  - domain: home.yourdomain.com
    provider: my-cloudflare
    datasource: my-public-ip

  - domain: vpn.yourdomain.com
    provider: my-cloudflare
    datasource: my-public-ip
    ipv6: true           # 这个域名同时更新 AAAA 记录

  - domain: blog.yourdomain.com
    enabled: false        # 暂时停用
    provider: my-cloudflare
    datasource: my-public-ip
```

每个域名独立按自己的周期运行，可以使用不同的数据源和 Provider。

## 下一步

- [工作原理](how-it-works.md) — 理解数据源、Provider 和分组策略
- [配置参考](../configuration/options.md) — 所有配置项详解
- [HTTP 数据源](../configuration/datasource/http.md) — HTTP 数据源完整参考
