This document was translated from Chinese by AI.

## Examples

Keep only IPv6 global unicast addresses:

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

Keep all IPv4 addresses and IPv6 global unicast addresses:

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
