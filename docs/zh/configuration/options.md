# 配置文件

Lightddns 配置文件的结构概览。完整示例可参考各子章节。

```yaml
log:
  level: info

datasources:
  - type: http
    name: data-http
    # ... HTTP 数据源配置

providers:
  - type: cloudflare
    name: prov-cf
    # ... Cloudflare Provider 配置

domains:
  - enabled: true
    domain: example.com
    provider: prov-cf
    datasource: data-http

services:
  - type: prometheus
    name: svc-metrics
    enabled: true
    # ... Prometheus 导出器配置
```

`datasources`、`providers` 和 `services` 都是带 `type` 的列表。如果
`name` 留空，Lightddns 在加载配置时会自动补上。

??? "name自动补全规则"
    ```yaml
    datasources:
        - type: http
        - type: command
    providers:
        - type: cloudflare
        - type: aliyun
    services:
        - type: ipserver
        - type: prometheus
    ```
    在自动补全后为
    ```yaml    
    datasources:
        - type: http
          name: datasource[0]
        - type: command
          name: datasource[1]
    providers:
        - type: cloudflare
          name: provider[0]
        - type: aliyun
          name: provider[1]
    services:
        - type: ipserver
          name: service[0]
        - type: prometheus
          name: service[1]
    ```

## Env 文件
CLI 全局参数 `--env-file` 用于设置 Lightddns 额外读取的环境变量来源。
文件需要符合 [标准dotenv](https://genkitlab.com/blog/dotenv-file-syntax/)

```env
# comment is allowed
PROVIDER_TOKEN=this_is_some_random_token_here
```

## 工作目录

CLI 全局参数 `-D` / `--workdir` 用于设置 Lightddns 的工作目录，默认是当前用户执行的文件夹。
设置工作目录会影响程序中所有相对路径形成的绝对路径。
程序中的绝对路径会按原样使用而不受 `-D` / `--workdir` 影响。
例如
```bash
$ lightddns -D /path_to_your_dir/lightddns -c config.yaml
```
将会读取 /path_to_your_dir/lightddns/config.yaml

```bash
lightddns -D /etc/lightddns  --env-file secrets.env run -c /etc/lightddns.yaml
```
将会读取 /etc/lightddns/secrets.env 与 /etc/lightddns.yaml

## `log`

全局日志配置。参见 [日志](log.md)。

## `datasources`

数据源列表。每个数据源负责获取当前主机的 IP 地址。

| Type                                 | 说明                                               |
|--------------------------------------|----------------------------------------------------|
| [`http`](datasource/http.md)         | 通过 HTTP 请求获取 IP。支持 JSON（jq）和正则提取。 |
| [`netlink`](datasource/netlink.md)   | 从本地网络接口读取 IP 地址。                       |
| [`command`](datasource/command.md)   | 通过执行 Shell 命令获取 IP 地址。                  |
| [`sum`](datasource/sum.md)           | 合并多个子数据源的 IP 地址。                       |
| [`failover`](datasource/failover.md) | 按优先级顺序查询子数据源，失败时自动切换。         |
| [`filter`](datasource/filter.md)     | 使用 CIDR 前缀规则过滤子数据源返回的 IP 地址。     |

## `providers`

服务提供方列表。每个 Provider 负责将 IP 地址比较并更新到对应的 DNS 服务商。

| Type                                       | 说明                           |
|--------------------------------------------|--------------------------------|
| [`cloudflare`](provider/cloudflare.md)     | 通过 Cloudflare API 更新。     |
| [`aliyun`](provider/aliyun.md)             | 通过阿里云解析（alidns）更新。 |
| [`tencentcloud`](provider/tencentcloud.md) | 通过腾讯云 DNSPod 更新。       |

## `domains`

域名记录列表。每个条目将域名、数据源和服务提供方绑定在一起，实现自动 DDNS 更新。

参见 [域名](domain.md)。

## `services`

后台 HTTP 服务列表。全部为可选项。

| Type                                  | 说明                        |
|---------------------------------------|-----------------------------|
| [`prometheus`](service/prometheus.md) | 导出内部 Prometheus 指标。  |
| [`ipserver`](service/ipserver.md)     | 一个轻量的 IP Echo 服务器。 |
