# 手册

Lightddns 由五个可组合的部分构成，共同完成一次完整的 DDNS 流程：

```
┌────────────┐   ┌──────────┐   ┌──────────┐
│ Datasource ├──▶│  Domain  ├──▶│ Provider │
└────────────┘   └──────────┘   └──────────┘
   获取当前 IP     调度更新周期     推送至 DNS 服务商

                  ┌──────────┐   ┌──────────┐
                  │   Log    │   │ Services │
                  └──────────┘   └──────────┘
                  全局日志     prometheus / ipserver
```

| 配置段 | 作用 | 详细文档 |
|---|---|---|
| `log` | 全局日志输出与级别 | [日志](../configuration/log.md) |
| `datasources` | 如何获取当前公网 IP — HTTP API、本机网卡、Shell 命令，或它们的组合 | [数据源](../configuration/options.md#datasources) |
| `providers` | 如何与 DNS 服务商通信 — Cloudflare、阿里云、腾讯云 | [服务提供方](../configuration/options.md#providers) |
| `domains` | 真正的绑定关系：用哪个 Provider 更新哪个域名、IP 从哪个数据源来、多久检查一次 | [域名](../configuration/domain.md) |
| `services` | 可选的后台 HTTP 服务 — Prometheus 指标导出、IP 回显 | [服务](../configuration/options.md#services) |

## 更新循环

每个启用的域名会以下面的顺序循环：

1. 向数据源询问当前 IP（按 `interval` 周期）。
2. 与 DNS 中已有记录进行 diff。
3. 仅当存在差异时，调用 Provider 增/改/删记录。

只有第 3 步会真正调用 DNS API，因此 IP 稳定时每个周期都是零 API 调用。

## 配置文件即 Go 模板

YAML 在解析前会先按 [Go template](https://pkg.go.dev/text/template) 渲染。`{{ .Env.NAME }}` 可以引用 OS 环境变量或工作目录下 `.env` 文件中的值：

```yaml
providers:
  - type: cloudflare
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"
```

这样可以把凭据从提交到 Git 的配置中隔离出去。

## 运行

```bash
# 持续运行
lightddns run -c lightddns.yaml

# 单次（执行一轮更新后退出）
lightddns run -c lightddns.yaml --once
```

## 示例

下方每个示例都是一份可直接使用的完整配置。把内容粘到 `lightddns.yaml`，设置好对应的环境变量后即可运行。

- [Cloudflare](cloudflare.md)
- [阿里云](aliyun.md)
- [腾讯云](tencentcloud.md)
- [分组数据源（sum / failover）](groups.md)
- [服务（Prometheus / IP 回显）](services.md)