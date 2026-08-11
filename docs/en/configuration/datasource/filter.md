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
    The filter datasource first queries its child datasources, then keeps only addresses matched by at least one rule. Rules are evaluated as OR: if any rule matches an address, that address is returned.

    Child datasources are queried with fast-fail behavior. If a child datasource returns an error, the filter datasource returns that error instead of returning a partial filtered result.

---

## `datasources`

A list of datasource names to read from. Each name must reference a datasource defined in the top-level `datasources` array.

```yaml
datasources:
  - data-netlink
  - data-http
```

---

## `rules`

Rules used to decide which addresses are kept. At least one rule is required.

```yaml
rules:
  - prefixes:
      - 203.0.113.0/24
```

Multiple rules are ORed:

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

CIDR prefixes matched by this rule.

```yaml
rules:
  - prefixes:
      - 0.0.0.0/0
      - ::/0
```

Use `0.0.0.0/0` to match all IPv4 addresses and `::/0` to match all IPv6 addresses.

If `prefixes` is empty or omitted, the rule matches every address before `invert` is applied.

---

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

This rule keeps addresses outside those three private IPv4 ranges.

Because rules are ORed, be careful when mixing inverted and non-inverted rules. An inverted broad rule can match addresses you expected another rule to exclude.

---

## Examples

Keep only IPv6 addresses from the global unicast range:

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

Keep all IPv4 addresses and global unicast IPv6 addresses:

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
