This document was translated from Chinese by AI.

# Filter

Filters IP addresses returned by one or more child datasources.

```yaml
# required
type: filter
name: data-filter
datasources:
  - data-http
rules:
  - prefixes:
      - 203.0.113.0/24
      - 2001:db8::/32
```

??? note "Behavior"
    The filter datasource first queries its child datasources, then keeps only addresses that match at least one rule. Multiple rules use OR semantics: if any rule matches an address, that address is returned.

## `datasources`

The names of child datasources to read. Each name must reference a datasource defined in the top-level `datasources` array.

```yaml
datasources:
  - data-netlink
  - data-http
```

## `rules`

Rules that determine which addresses are kept. At least one rule is required.

```yaml
rules:
  - prefixes:
      - 203.0.113.0/24
```

Multiple rules use OR semantics:

```yaml
rules:
  - prefixes:
      - 203.0.113.0/24
  - prefixes:
      - 2001:db8::/32
```

This keeps addresses in either `203.0.113.0/24` or `2001:db8::/32`.

---

## `rules[].prefixes`

The list of CIDR prefixes matched by this rule.

```yaml
rules:
  - prefixes:
      - 0.0.0.0/0
      - ::/0
```

Use `0.0.0.0/0` to match all IPv4 addresses and `::/0` to match all IPv6 addresses.

If `prefixes` is empty or omitted, the rule matches all addresses before `invert` is applied.

## `rules[].invert`

Inverts the result of this rule.

```yaml
rules:
  - prefixes:
      - 10.0.0.0/8
      - 172.16.0.0/12
      - 192.168.0.0/16
    invert: true
```

This rule keeps addresses outside these three private IPv4 ranges.

Because multiple rules use OR semantics, take particular care when mixing inverted and ordinary rules. An inverted rule with a broad range may match addresses that you intended another rule to exclude.
