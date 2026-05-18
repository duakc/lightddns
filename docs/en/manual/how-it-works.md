# How It Works

This page explains the concepts behind Lightddns in plain language. You don't need to be a programmer to understand how it works.

## The Big Picture

Lightddns does one job: **keep your DNS records pointing at your current IP address**. To do this, it uses three building blocks:

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Datasource  │────▶│    Domain    │────▶│   Provider   │
│ (finds IP)   │     │ (ties them   │     │ (updates DNS)│
│              │     │  together)   │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
```

1. **Datasource** — answers "what is my current IP address?"
2. **Provider** — talks to your DNS hosting service to update records
3. **Domain** — connects a domain name to a datasource and provider, and controls how often to check

## Datasources

A datasource is how Lightddns figures out your public IP address. There are four types:

### HTTP Datasource

Asks a web service what your IP is. Many free services exist for this purpose.

```
Your Machine ──HTTP──▶ ipify.org ──▶ "Your IP is 203.0.113.5"
```

The response can be plain text, JSON, or any format you can parse with a regular expression. The [HTTP datasource reference](../configuration/datasource/http.md) has full details.

**When to use:** This is the most common and recommended method for home users behind a router.

### Command Datasource

Runs a shell command or script. Whatever the command prints is scanned for IP addresses.

```yaml
cmd:
  ipv4: "dig +short myip.opendns.com @resolver1.opendns.com"
```

**When to use:** When you have a specific script, need custom logic, or want to use a command-line tool like `dig` to discover your IP.

### Netlink Datasource

Reads IP addresses directly from a network interface on your machine.

**When to use:** When your machine has a public IP directly assigned to its network interface (common on VPS, rare on home connections behind NAT).

### Group Datasources

Groups combine multiple datasources together. Two strategies are available:

**Sum** — collect IPs from all child datasources and merge them. If you configure both an IPv4 HTTP datasource and a separate IPv6 one, a `sum` group can combine both results.

```yaml
- type: sum
  name: all-ips
  datasources:
    - ipv4-http
    - ipv6-http
```

**Failover** — try datasources in order, moving to the next one if a datasource fails. This is useful when you have a primary and backup IP lookup method.

```yaml
- type: failover
  name: reliable-ip
  datasources:
    - primary-http     # tried first
    - backup-command   # used if primary fails
```

!!! tip "Dependency ordering"
    Group datasources reference other datasources by name. Lightddns automatically figures out the correct initialization order — you don't need to list them in any particular order in the config file.

## Providers

A provider is the DNS hosting service that actually stores your DNS records. Currently, Lightddns supports **Cloudflare**.

### How Cloudflare Integration Works

1. Lightddns queries the Cloudflare API to find your domain's Zone ID
2. It lists the existing DNS records for that zone
3. It compares (diffs) your desired IP against what's currently in Cloudflare
4. If there's a difference, it creates, updates, or deletes records to match

Lightddns only makes API calls when the IP has actually changed. Between changes, it just checks your IP locally at the configured interval.

!!! info "Proxy (Orange Cloud)"
    The Cloudflare provider supports the `proxy` option. When enabled, traffic to your domain goes through Cloudflare's network (hiding your real IP). When disabled, DNS responds with your actual IP. The default is `false` (no proxy).

## Domains

A domain ties everything together. Each domain entry in the config defines:

| Setting | What it does | Default |
|---|---|---|
| `domain` | The DNS name to update | *(required)* |
| `provider` | Which provider to use | *(required)* |
| `datasource` | Where to get the IP from | *(required)* |
| `interval` | How often to check for IP changes | `30s` |
| `timeout` | Maximum time for each check + update | `15s` |
| `ttl` | DNS record TTL in seconds | Provider default |
| `ipv4` | Whether to update the A record (IPv4) | `true` |
| `ipv6` | Whether to update the AAAA record (IPv6) | `true` |
| `enabled` | Whether this domain is active | `true` |

## The Update Loop

Here's what happens every `interval` for each domain:

1. **Check IP** — Lightddns asks the datasource for your current IP
2. **Compare** — The provider checks if this IP differs from what's currently in DNS
3. **Update** — If the IP changed (or the record doesn't exist), the provider creates or updates the DNS record
4. **Skip** — If nothing changed, Lightddns does nothing and waits for the next interval

This means your DNS records are only touched when something actually changes. No unnecessary API calls, no rate-limiting issues.

## Configuration Files as Templates

Lightddns config files are processed as [Go templates](https://pkg.go.dev/text/template) before being parsed as YAML. This lets you use `{{.Env.VARIABLE}}` to reference environment variables:

```yaml
providers:
  - type: cloudflare
    name: cf
    token: "{{.Env.CF_API_TOKEN}}"
```

Environment variables are loaded from two sources (merged together):
1. Your actual OS environment (`export CF_API_TOKEN=...`)
2. A `.env` file in the current working directory

This keeps secrets like API tokens out of your config file — you can safely commit the config to git or share it without exposing credentials.

## Running Modes

Lightddns has two running modes:

**Continuous mode** (default):
```bash
lightddns run -c lightddns.yaml
```
Runs forever, checking each domain on its configured interval. Press `Ctrl+C` to stop.

**One-shot mode:**
```bash
lightddns run -c lightddns.yaml --once
```
Checks each domain once and exits. Useful for testing, cron jobs, or running Lightddns through an external scheduler.

## Logging

By default, Lightddns logs to your terminal (stdout). You can control this in the config:

```yaml
log:
  level: info        # debug, info, warn, error
  output: ""         # empty = stdout; set a path to write to a file
```

For troubleshooting, set `level: debug` to see what's happening at each step.

## Next Steps

- [Configuration Reference](../configuration/options.md) — every option documented in detail
- [HTTP Datasource Reference](../configuration/datasource/http.md) — all HTTP datasource settings
- [Domain Configuration](../configuration/domain.md) — domain-specific options
