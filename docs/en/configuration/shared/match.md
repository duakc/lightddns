This document was translated from Chinese by AI.

# MatchOption

Shared rules for extracting IP addresses from command output or HTTP response bodies.

```yaml
match:
  jq: ".ip"
  regex: "IP:\\s+(\\S+)"
```

When `match` is omitted, Lightddns still attempts plain-text parsing.

## `jq`

The query language described in the [jq manual](https://jqlang.org/manual/).
The current implementation uses [gojq](https://github.com/itchyny/gojq), a Go implementation.
Expressions should be understood according to the [jq manual](https://jqlang.org/manual/), but syntax parsing and runtime compatibility are determined by [`gojq`](https://github.com/itchyny/gojq).

```yaml
match:
  jq: ".ip"
```

An expression can produce one or more values:

```yaml
match:
  jq: ".addresses[]"
```

Only string results are parsed as IP addresses. Objects, arrays, numbers, booleans, and `null` are ignored. If a string result is not a valid IP literal, that jq extraction fails.

## `regex`

Uses a regular expression to extract IP addresses.

```yaml
match:
  regex: "Current IP:\\s+(\\S+)"
```

The expression must contain a capture group. After each regular-expression match, Lightddns parses only the first capture group as an IP address, not the full match.

```yaml
# Correct: the first capture group contains only the address.
match:
  regex: "IP=(\\S+)"

# Incorrect: there is no capture group.
match:
  regex: "IP=\\S+"
```

## Plain Text

When the preceding rules return no addresses, Lightddns trims leading and trailing whitespace, splits the input on Unicode whitespace, and parses every segment as an IP address.

The following output can be parsed directly:

```text
203.0.113.10
 some_garbage_but_has_space_near
2001:db8::10
```

## Order

Datasources that use the default matcher, such as `command`, extract addresses in this order:

1. If `jq` is configured, parse the input as JSON and run the jq expression. Return immediately only if jq returns at least one IP address.
2. If `regex` is configured and jq returned no addresses, parse the first capture group of every regular-expression match. Return immediately only if the regular expression returns at least one IP address.
3. If no address has been found, use plain-text parsing.
4. If all enabled strategies return no addresses, return the collected extraction errors.

The HTTP datasource also considers the response `Content-Type`:

1. If the response `Content-Type` is JSON and `match.jq` is configured, HTTP uses jq only; it does not fall back to a regular expression or plain text.
2. Otherwise, HTTP tries `match.regex` first, then plain-text parsing.

Use `match.jq` for JSON endpoints.
Use `match.regex` for text pages that contain explanatory text.
If the response body contains only one or more whitespace-separated IP literals, `match` can be omitted.
