## 示例

只保留 IPv6 全局单播地址：

```yaml
datasources:
  - type: filter
    name: data-global-v6
    datasources:
      - data-all
    rules:
      - prefixes:
          - 2000::/3
```

保留所有 IPv4 地址和 IPv6 全局单播地址：

```yaml
datasources:
  - type: filter
    name: data-v4-and-global-v6
    datasources:
      - data-all
    rules:
      - prefixes:
          - 0.0.0.0/0
      - prefixes:
          - 2000::/3
```
