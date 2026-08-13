This document was translated from Chinese by AI.

# Sum

Merges IP addresses from multiple child datasources into one result set.

```yaml
# required
type: sum
name: data-sum
datasources:
  - data-http
  - data-netlink
```

??? note "Behavior"
    Child datasources are queried in order, and the resulting IP addresses are concatenated.
    If any child datasource returns an error, the sum datasource immediately returns that error.
    To retry until a query succeeds, use [Failover](failover.md).

## `datasources`

The list of datasource names to merge. Each name must reference a datasource defined in the configuration file's `datasources` array. At least one datasource is required.

```yaml
datasources:
  - data-http
  - data-cmd
```
