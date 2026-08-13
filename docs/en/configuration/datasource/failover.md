This document was translated from Chinese by AI.

# Failover

Queries child datasources in order and automatically switches to the next one on failure.

```yaml
# required
type: failover
name: data-failover
datasources:
  - data-netlink
  - data-http
```

??? note "Behavior"
    Uses a **sticky success** strategy: once a datasource succeeds, subsequent queries start from that datasource to avoid unnecessary retries. If it fails, traversal continues forward through the list to try the next datasource.

    The maximum number of traversals is half the list length, limiting retry time. If every attempted datasource fails, the group returns an error.

---

## `datasources`

A list of datasource names in priority order. The first is tried first; on failure, the second is tried, and so on. Each name must reference a datasource defined in the configuration file's `datasources` array. At least one datasource is required.

```yaml
datasources:
  - data-netlink     # Tried first (local and fast)
  - data-http        # Fallback (remote API)
```
