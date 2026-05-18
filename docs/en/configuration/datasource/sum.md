# Sum

Merges IP addresses from multiple child datasources into a single result set.

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

A list of datasource names to merge. Each name must reference a datasource defined in the `datasources` array of the config file. At least one datasource is required.

```yaml
datasources:
  - data-http
  - data-cmd
```

---

## `fastFail`

When `true`, the group fails immediately if any child datasource returns an error. When `false` (default), errors from individual datasources are collected; the group only fails if **all** datasources fail or return empty results.

```yaml
fastFail: true
```
