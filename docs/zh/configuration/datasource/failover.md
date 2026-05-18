# Failover

按顺序查询子数据源，失败时自动切换到下一个。

```yaml
# required
type: failover
name: data-failover
datasources:
  - data-netlink
  - data-http
```

??? note "行为说明"
    采用**粘性成功**策略：一旦某个数据源成功，后续查询会从该数据源开始，避免不必要的重试。若失败，则向后遍历列表尝试下一个。

    遍历次数上限为列表长度的一半，用于限制重试时间。若所有尝试的数据源均失败，组返回错误。

---

## `datasources`

按优先级排列的数据源名称列表。优先尝试第一个，失败则切换到第二个，以此类推。每个名称必须引用配置文件中 `datasources` 数组内已定义的数据源。至少需要一个数据源。

```yaml
datasources:
  - data-netlink     # 优先尝试（本地，速度快）
  - data-http        # 备用方案（远程 API）
```

---
