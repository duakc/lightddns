# 分组数据源

`sum` 和 `failover` 都是以其它数据源作为输入的"组合数据源"。它们让你能在不改变域名引用方式的前提下，把多个 IP 发现策略组合起来。

## Sum

收集所有子数据源的 IP 并合并结果。常用于 IPv4 与 IPv6 来自不同接口的场景。

```yaml
datasources:
  - type: http
    name: data-v4
    url: https://api.ipify.org
    dialStrategy: ipv4_only

  - type: http
    name: data-v6
    url: https://api6.ipify.org
    dialStrategy: ipv6_only

  - type: sum
    name: data-all
    datasources:
      - data-v4
      - data-v6

providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-all
    ipv4: true
    ipv6: true
```

默认情况下 `sum` 仅当 **所有** 子数据源都失败时才返回错误。设 `fastFail: true` 后任意子数据源出错都会立即中止。

## Failover

按顺序尝试子数据源，直到某一个成功为止。常用于"本地优先、远端兜底"的场景。

```yaml
datasources:
  - type: netlink
    name: data-nic
    ifName: eth0

  - type: http
    name: data-http
    url: https://api.ip.sb/ip

  - type: failover
    name: data-fo
    datasources:
      - data-nic    # 优先尝试本地网卡
      - data-http   # 失败则回退到公开 API

providers:
  - type: cloudflare
    name: prov-cf
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"

domains:
  - enabled: true
    domain: home.example.com
    provider: prov-cf
    datasource: data-fo
```

一旦某个子数据源成功，下一轮会直接从它开始 — 即使写很长的回退链也不会带来额外开销。