# Sum

将多个子数据源的 IP 地址合并为一个结果集。

```yaml
# required
type: sum
name: data-sum
datasources:
  - data-http
  - data-netlink

# optional
fastFail: false
```

## `datasources`

要合并的数据源名称列表。每个名称必须引用配置文件中 `datasources` 数组内已定义的数据源。至少需要一个数据源。

```yaml
datasources:
  - data-http
  - data-cmd
```

---

## `fastFail`

设为 `true` 时，任意子数据源返回错误，整个组立即失败。设为 `false`（默认）时，收集各数据源的错误；仅当**所有**数据源均失败或返回空结果时才报错。

```yaml
fastFail: true
```
