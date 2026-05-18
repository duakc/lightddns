# Failover

Queries child datasources in order, falling back to the next one on failure.

```yaml
# required
type: failover
name: data-failover
datasources:
  - data-netlink
  - data-http
```

??? note "Behavior"
    Uses a **sticky success** strategy: once a datasource succeeds, subsequent queries start from that same datasource to avoid unnecessary retries. If it fails, the group walks forward through the list.

    The walk is capped at half the list length to bound retry time. If all attempted datasources fail, the group returns an error.

---

## `datasources`

A prioritized list of datasource names. The first datasource is tried first; if it fails, the next one is tried, and so on. Each name must reference a datasource defined in the `datasources` array of the config file. At least one datasource is required.

```yaml
datasources:
  - data-netlink     # try first (fast, local)
  - data-http        # fallback (remote API)
```

---
