# Sum

将多个子数据源的 IP 地址合并为一个结果集。

```yaml
# required
type: sum
name: data-sum
datasources:
  - data-http
  - data-netlink
```

??? note "行为说明"
    子数据源会按顺序查询，并把得到的 IP 地址拼接成一个结果。当前实现使用快速失败行为：任意子数据源返回错误时，sum 数据源会立即返回该错误。

## `datasources`

要合并的数据源名称列表。每个名称必须引用配置文件中 `datasources` 数组内已定义的数据源。至少需要一个数据源。

```yaml
datasources:
  - data-http
  - data-cmd
```
