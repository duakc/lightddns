# Sum

Merges IP addresses from multiple child datasources into a single result set.

```yaml
# required
type: sum
name: data-sum
datasources:
  - data-http
  - data-netlink
```

??? note "Behavior"
    Child datasources are queried in order and their IP addresses are concatenated. The current implementation uses fast-fail behavior: if any child datasource returns an error, the sum datasource returns that error immediately.

## `datasources`

A list of datasource names to merge. Each name must reference a datasource defined in the `datasources` array of the config file. At least one datasource is required.

```yaml
datasources:
  - data-http
  - data-cmd
```
