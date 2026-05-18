# Lightddns — Lightweight Dynamic DNS

Lightddns keeps your domain name pointed at your home IP address, automatically. Whenever your ISP changes your IP, Lightddns detects the change and updates your DNS records in seconds.

## What is DDNS?

Most home internet connections get a **dynamic IP address** — your ISP can change it at any time. If you run a NAS, security camera, game server, or home VPN at home, you need a stable domain name to reach it from outside. DDNS (Dynamic DNS) solves this: a small program runs on your machine, watches your public IP, and updates your domain's DNS records whenever the IP changes.

## Why Lightddns?

- **Simple YAML config** — one file, easy to read and edit
- **Works with your existing domains** — managed through Cloudflare (more providers coming)
- **Multiple IP sources** — HTTP APIs, local network interfaces, or shell commands
- **Dual-stack support** — handles both IPv4 and IPv6 for each domain
- **Smart groups** — combine datasources (sum) or create failover chains
- **Low resource usage** — a single, small binary with minimal footprint
- **Template configs** — use `{{.Env.VAR}}` to pull secrets from environment variables

## Quick Start

### 1. Download

Get the latest binary from the [GitHub Releases](https://github.com/duakc/lightddns/releases) page. Choose the one for your operating system and architecture.

### 2. Create a config file

Create a file named `lightddns.yaml`:

```yaml
datasources:
  - type: http
    name: my-ip
    url: https://api64.ipify.org?format=json
    json: ".ip"

providers:
  - type: cloudflare
    name: cf
    token: "{{.Env.CF_API_TOKEN}}"

domains:
  - domain: home.example.com
    provider: cf
    datasource: my-ip
```

### 3. Run it

```bash
lightddns run -c lightddns.yaml
```

Lightddns will check your IP every 30 seconds and update the DNS record whenever it changes. Use `--once` to run a single check and exit.

!!! tip "Environment Variables"
    Store your Cloudflare API token in a `.env` file or export it as `CF_API_TOKEN`. The `{{.Env.CF_API_TOKEN}}` in the config will be replaced automatically.

## How It Works

Lightddns connects three things together:

1. **Datasource** — discovers your current public IP address. Options include querying HTTP APIs (like ipify.org), reading from a local network interface, or running a custom script.
2. **Provider** — pushes DNS updates to your DNS hosting service (currently Cloudflare).
3. **Domain** — binds a domain name to a datasource + provider, and sets the update schedule.

Each domain runs independently, checks its datasource on a configurable interval, and calls the provider to update DNS records when the IP has changed.

## Next Steps

- [Installation Guide](manual/installation.md) — detailed install steps for all platforms
- [Getting Started](manual/getting-started.md) — walkthrough of your first setup
- [How It Works](manual/how-it-works.md) — deeper explanation for curious users
- [Configuration Reference](configuration/options.md) — every option explained
