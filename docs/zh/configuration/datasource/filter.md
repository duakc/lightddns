# Filter

过滤一个或多个子数据源返回的 IP 地址。

```yaml
# required
type: filter
name: data-filter
datasources:
  - data-http
rules:
  - prefixes:
      - 203.0.113.0/24
      - 2001:db8::/32
```

??? note "行为说明"
    filter 数据源会先查询子数据源，再只保留至少命中一条规则的地址。多条规则之间是 OR 关系：任意规则匹配某个地址，该地址就会被返回。


## `datasources`

要读取的子数据源名称列表。每个名称必须引用顶层 `datasources` 数组中已经定义的数据源。

```yaml
datasources:
  - data-netlink
  - data-http
```


## `rules`

用于决定保留哪些地址的规则。至少需要一条规则。

```yaml
rules:
  - prefixes:
      - 203.0.113.0/24
```

多条规则之间是 OR 关系：

```yaml
rules:
  - prefixes:
      - 203.0.113.0/24
  - prefixes:
      - 2001:db8::/32
```

这会保留位于 `203.0.113.0/24` 或 `2001:db8::/32` 的地址。

---

## `rules[].prefixes`

本条规则匹配的 CIDR 前缀列表。

```yaml
rules:
  - prefixes:
      - 0.0.0.0/0
      - ::/0
```

使用 `0.0.0.0/0` 匹配所有 IPv4 地址，使用 `::/0` 匹配所有 IPv6 地址。

如果 `prefixes` 留空或省略，在应用 `invert` 之前，这条规则会匹配所有地址。


## `rules[].invert`

反转本条规则的匹配结果。

```yaml
rules:
  - prefixes:
      - 10.0.0.0/8
      - 172.16.0.0/12
      - 192.168.0.0/16
    invert: true
```

这条规则会保留不在这三个私有 IPv4 网段内的地址。

由于多条规则之间是 OR 关系，混用反转规则和普通规则时要特别小心。范围很大的反转规则可能会匹配到你原本希望由其它规则排除的地址。
