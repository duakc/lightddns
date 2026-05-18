# 工作原理

本页用通俗的语言解释 Lightddns 的工作机制。不需要编程知识即可理解。

## 总览

Lightddns 只做一件事：**让 DNS 记录始终指向你当前的 IP 地址**。为此，它使用三个核心组件：

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   数据源      │────▶│     域名      │────▶│   服务提供方   │
│  (获取 IP)    │     │  (绑定关系)   │     │  (更新 DNS)   │
│              │     │              │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
```

1. **数据源（Datasource）**——回答"我当前的 IP 是什么？"
2. **服务提供方（Provider）**——与 DNS 服务商通信，更新记录
3. **域名（Domain）**——将域名与数据源和 Provider 绑定，并控制检查频率

## 数据源

数据源告诉 Lightddns 你的公网 IP 地址是多少。有四种类型：

### HTTP 数据源

向一个 web 服务查询你的 IP。有很多免费服务可用。

```
你的设备 ──HTTP──▶ ipify.org ──▶ "你的 IP 是 203.0.113.5"
```

响应可以是纯文本、JSON 或其他可用正则表达式解析的格式。详见 [HTTP 数据源参考](../configuration/datasource/http.md)。

**适用场景：** 这是家庭用户最常用的方式，适合路由器后的设备。

### 命令数据源

执行一条 Shell 命令或脚本。命令的输出会被自动扫描提取 IP 地址。

```yaml
cmd:
  ipv4: "dig +short myip.opendns.com @resolver1.opendns.com"
```

**适用场景：** 当你已有特定脚本、需要自定义逻辑，或想用 `dig` 等命令行工具获取 IP。

### 网卡数据源

直接读取本机网络接口上的 IP 地址。

**适用场景：** 当你的机器网卡直接绑定公网 IP（常见于 VPS，家庭 NAT 网络下很少见）。

### 分组数据源

分组将多个数据源组合在一起。支持两种策略：

**Sum（求和）**——收集所有子数据源的 IP 并合并。如果你分别配置了 IPv4 HTTP 数据源和 IPv6 HTTP 数据源，用 `sum` 分组可以把两者结果合并。

```yaml
- type: sum
  name: all-ips
  datasources:
    - ipv4-http
    - ipv6-http
```

**Failover（故障转移）**——按顺序尝试数据源，前面失败了就换下一个。适用于为主数据源准备备份方式的场景。

```yaml
- type: failover
  name: reliable-ip
  datasources:
    - primary-http     # 优先使用
    - backup-command   # 主数据源失败时启用
```

!!! tip "依赖排序"
    分组数据源通过名称引用其他数据源。Lightddns 会自动计算正确的初始化顺序——配置文件中的数据源不必按特定顺序排列。

## 服务提供方

Provider 是实际存储 DNS 记录的服务商。目前 Lightddns 支持 **Cloudflare**。

### Cloudflare 集成流程

1. Lightddns 查询 Cloudflare API 获取域名的 Zone ID
2. 列出该 Zone 下已有的 DNS 记录
3. 对比（diff）你想设置的 IP 和 Cloudflare 中当前的记录
4. 如有差异，自动创建、更新或删除记录

Lightddns 只在 IP 实际发生变化时才调用 Cloudflare API。在变化之间，仅在本地检查 IP。

!!! info "代理（橙色云朵）"
    Cloudflare Provider 支持 `proxy` 选项。开启后，访问域名的流量走 Cloudflare 网络（隐藏真实 IP）。关闭时，DNS 直接返回真实 IP。默认为 `false`（不代理）。

## 域名

Domain 把一切聚合在一起。配置文件中的每个域名条目定义了：

| 设置 | 作用 | 默认值 |
|---|---|---|
| `domain` | 要更新的 DNS 名称 | *(必填)* |
| `provider` | 使用哪个 Provider | *(必填)* |
| `datasource` | 从哪获取 IP | *(必填)* |
| `interval` | 多久检查一次 IP 变化 | `30s` |
| `timeout` | 每次检查+更新的最大耗时 | `15s` |
| `ttl` | DNS 记录 TTL，单位秒 | 由 Provider 决定 |
| `ipv4` | 是否更新 A 记录（IPv4） | `true` |
| `ipv6` | 是否更新 AAAA 记录（IPv6） | `true` |
| `enabled` | 是否启用该域名 | `true` |

## 更新流程

每个 `interval`，每个域名都会经历以下步骤：

1. **检查 IP**——Lightddns 向数据源查询当前 IP
2. **对比**——Provider 检查此 IP 是否与 DNS 中现有记录不同
3. **更新**——如果 IP 变了（或记录不存在），Provider 创建或更新 DNS 记录
4. **跳过**——如果一切没变，Lightddns 不做任何操作，等待下一个周期

这意味着你的 DNS 记录只在实际发生变化时才会被修改。没有多余的 API 调用，不会有速率限制问题。

## 配置文件模板

Lightddns 配置文件在解析 YAML 前先经过 [Go 模板](https://pkg.go.dev/text/template)处理。你可以在配置中使用 `{{.Env.VARIABLE}}` 引用环境变量：

```yaml
providers:
  - type: cloudflare
    name: cf
    token: "{{.Env.CF_API_TOKEN}}"
```

环境变量从两个来源合并加载：
1. 操作系统的环境变量（`export CF_API_TOKEN=...`）
2. 当前工作目录下的 `.env` 文件

这样 API 令牌等敏感信息就不会暴露在配置文件中——你可以安全地将配置文件提交到 git 或分享给他人。

## 运行模式

Lightddns 有两种运行模式：

**持续运行**（默认）：
```bash
lightddns run -c lightddns.yaml
```
持续运行，按各域名设定的间隔反复检查。按 `Ctrl+C` 停止。

**单次运行：**
```bash
lightddns run -c lightddns.yaml --once
```
每个域名检查一次后退出。适用于测试、cron 定时任务、或通过外部调度器控制运行节奏。

## 日志

默认情况下，Lightddns 将日志输出到终端（stdout）。可以在配置中调整：

```yaml
log:
  level: info        # debug、info、warn、error
  output: ""         # 为空则输出到 stdout；填写路径则写入文件
```

排查问题时，设置 `level: debug` 可以查看每一步的详细信息。

## 下一步

- [配置参考](../configuration/options.md) — 所有配置项详细说明
- [HTTP 数据源参考](../configuration/datasource/http.md) — HTTP 数据源全部设置
- [域名配置](../configuration/domain.md) — 域名相关配置项
