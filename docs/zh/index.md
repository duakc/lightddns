# Lightddns — 轻量级动态 DNS

Lightddns 让你的域名始终指向你当前的 IP 地址，完全自动。每当运营商更换你的公网 IP，Lightddns 会在几秒内检测到变化并更新 DNS 记录。

## 什么是 DDNS？

大多数家庭宽带获取的是**动态 IP 地址**——运营商随时可能更换。如果你在家里搭建了 NAS、监控摄像头、游戏服务器或 VPN，就需要一个固定的域名才能从外面访问。DDNS（动态 DNS）就是解决方案：在你的设备上运行一个小程序，持续监测公网 IP，IP 一变就立刻更新域名解析记录。

## 为什么选 Lightddns？

- **简洁的 YAML 配置**——一个文件，清晰易读
- **多家 DNS 服务商**——Cloudflare、阿里云、腾讯云
- **多种 IP 获取方式**——HTTP API、本地网卡、或自定义脚本
- **双栈支持**——每个域名可同时管理 IPv4 和 IPv6
- **智能分组**——合并数据源（`sum`）或串成失败回退（`failover`）
- **自带指标**——可选启用 Prometheus 导出器，同一个二进制内
- **资源占用极低**——单个小巧的二进制文件
- **模板配置**——通过 `{{ .Env.VAR }}` 从环境变量注入密钥

## 快速开始

### 1. 下载

从 [GitHub Releases](https://github.com/duakc/lightddns/releases) 页面获取对应你系统的最新版本。

### 2. 创建配置文件

新建一个 `lightddns.yaml`：

```yaml
datasources:
  - type: http
    name: data-http
    url: https://api.ip.sb/ip
    dialStrategy: ipv4_only

providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-http
    ipv4: true
```

### 3. 运行

```bash
lightddns run -c example.yaml
```

Lightddns 默认每 30 秒检查一次 IP，仅在变化时更新 DNS。加 `--once` 参数可以只执行一次后退出。

!!! tip "环境变量"
    将 Cloudflare API 令牌存入 `.env` 文件或导出为环境变量 `CLOUDFLARE_TOKEN`。配置文件中的 `{{ .Env.CLOUDFLARE_TOKEN }}` 会自动替换。

## 工作原理

Lightddns 把五个部分组合在一起：

1. **数据源（Datasource）**——获取你当前的公网 IP。可以通过 HTTP API 查询（如 ipify.org）、读取本机网卡地址、或执行自定义脚本。
2. **服务提供方（Provider）**——向 DNS 服务商（Cloudflare、阿里云、腾讯云）推送 DNS 更新。
3. **域名（Domain）**——将域名绑定到数据源和 Provider，并设定更新周期。
4. **日志（Log）**——全局日志输出与级别。
5. **服务（Services）**——可选的后台 HTTP 服务，如 Prometheus 导出器或 IP 回显服务。

每个域名独立运行，按设定的间隔检查 IP，仅在 IP 变化时才会调用服务商接口更新 DNS 记录。

## 下一步

- [手册](manual/index.md) — 架构总览与即用即跑的示例
- [配置参考](configuration/options.md) — 所有配置项详解