# MatchOption

Shared IP extraction rules for datasources that parse command output or HTTP response bodies. These fields live under the `match:` key.

```yaml
match:
  jq: ".ip"
  regex: "IP:\\s+(\\S+)"
```

Both fields are optional. If `match` is omitted, Lightddns still tries the plain-text fallback.

## `jq`

`jq` is the query language described by the [jq manual](https://jqlang.org/manual/). Lightddns does not implement the language itself; it currently evaluates this field with the Go implementation [gojq](https://github.com/itchyny/gojq). In practice, write jq filters as you would for jq, but parsing and runtime compatibility are determined by `gojq`.

```yaml
match:
  jq: ".ip"
```

The expression may emit one or more values:

```yaml
match:
  jq: ".addresses[]"
```

Only string results are parsed as IP addresses. Non-string jq results are ignored. If a string result is not a valid IP literal, that jq extraction fails.

## `regex`

Extracts IP addresses with a regular expression.

```yaml
match:
  regex: "Current IP:\\s+(\\S+)"
```

The expression must contain a capture group. For every regex match, Lightddns parses the first capture group as the IP address. The full match is not parsed.

```yaml
# Good: the first capture group is only the address.
match:
  regex: "IP=(\\S+)"

# Wrong: there is no capture group for Lightddns to parse.
match:
  regex: "IP=\\S+"
```

## Plain Text

When no earlier rule returns an address, Lightddns trims the input, splits it by Unicode whitespace, and parses each token as an IP address.

This works:

```text
203.0.113.10
2001:db8::10
```

This needs `regex` or `jq`:

```text
ip=203.0.113.10
"203.0.113.10"
```

## Order

Datasources using the default matcher, such as `command`, use this order:

1. If `jq` is set, parse the input as JSON and run the jq expression. Return immediately only when jq returns at least one IP address.
2. If `regex` is set and jq did not return an address, parse the first capture group from every regex match. Return immediately only when regex returns at least one IP address.
3. If no address has been found, use plain-text parsing.
4. If every enabled strategy fails to produce an address, return the collected extraction errors.

The HTTP datasource adds one rule because HTTP responses have a `Content-Type` header:

1. If the response `Content-Type` is JSON and `match.jq` is set, HTTP uses jq only. It does not fall back to regex or plain text.
2. Otherwise, HTTP tries `match.regex` when set, then plain text.

Use `match.jq` for JSON endpoints. Use `match.regex` for text pages that include labels or extra words. Omit `match` when the body is only one or more whitespace-separated IP literals.
