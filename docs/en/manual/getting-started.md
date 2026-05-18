# Getting Started

This guide walks you through setting up Lightddns to automatically update a domain with your home IP address. It takes about 5 minutes.

## What You Need

- A domain name managed by **Cloudflare** (other providers coming soon)
- A Cloudflare **API token** with DNS edit permission
- The Lightddns binary installed (see [Installation](installation.md))

### Getting a Cloudflare API Token

1. Log into the [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Go to **My Profile** → **API Tokens** (or [click here](https://dash.cloudflare.com/profile/api-tokens))
3. Click **Create Token**
4. Use the **Edit zone DNS** template, or create a custom token with these permissions:
    - `Zone` — `DNS` — `Edit`
    - `Zone` — `Zone` — `Read`
5. Under **Zone Resources**, select the domain you want to manage
6. Click **Continue to summary** → **Create Token**
7. Copy the token — you'll need it in the config file

## Step 1: Create the Config File

Create a new file. The default location `/etc/lightddns/lightddns.yaml` works well, but any path is fine.

```yaml
# lightddns.yaml

datasources:
  # Discovers your public IPv4 address using a free HTTP API
  - type: http
    name: my-public-ip
    url: https://api64.ipify.org?format=json
    json: ".ip"

providers:
  # Connects to Cloudflare to update DNS records
  - type: cloudflare
    name: my-cloudflare
    token: "{{.Env.CF_API_TOKEN}}"

domains:
  # Ties it all together: check IP, update home.example.com on Cloudflare
  - domain: home.yourdomain.com
    provider: my-cloudflare
    datasource: my-public-ip

    # Optional: customize the update behavior
    interval: 5m       # check every 5 minutes
    ttl: 300           # DNS TTL of 5 minutes
    ipv4: true         # update the A record
    ipv6: false        # don't update the AAAA record
```

Replace `home.yourdomain.com` with your actual domain.

!!! warning "The domain must already exist in Cloudflare"
    Lightddns only updates existing DNS records. Before running, log into Cloudflare and create a DNS **A** record for `home.yourdomain.com` pointing to any IP address (e.g., `1.2.3.4`). Lightddns will update it to your real IP on the first run.

## Step 2: Set Your API Token

Create a `.env` file next to your config (or export the variable directly):

```bash
# .env
CF_API_TOKEN=your-cloudflare-api-token-here
```

The `{{.Env.CF_API_TOKEN}}` in the config file will be replaced with the value from your environment or `.env` file. This keeps secrets out of the config file itself.

## Step 3: Test It

Run Lightddns once to verify everything works:

```bash
lightddns run -c lightddns.yaml --once
```

If the log output looks like this, it's working:

```
INFO  main  ipv4 ip updated  domain=home.yourdomain.com  ip=203.0.113.5
```

Check your Cloudflare dashboard — the DNS record should now show your actual public IP.

## Step 4: Run It Permanently

Remove the `--once` flag to run continuously:

```bash
lightddns run -c lightddns.yaml
```

Lightddns will check your IP on the interval you configured and update DNS whenever it changes. Press `Ctrl+C` to stop.

For production use, set it up as a [systemd service](installation.md#running-as-a-background-service) so it starts automatically and runs in the background.

## Customizing IP Sources

The HTTP datasource above uses [ipify.org](https://www.ipify.org) — a free, reliable IP lookup service. You can use any service that returns your IP address.

### Plain-Text APIs

Some services return a plain-text IP address:

```yaml
datasources:
  - type: http
    name: ipv4-http
    url: https://api.ip.sb/ip
    # no json/regex needed — the raw response body is used as the IP
```

### Dual-Stack (IPv4 + IPv6)

To get both IPv4 and IPv6 addresses, configure separate jq expressions and use `prefer_ipv6` or the default dial strategy:

```yaml
datasources:
  - type: http
    name: dualstack-http
    url: https://api64.ipify.org?format=json
    json:
      ipv4: ".ip"
    # For IPv6, you need a service that returns your IPv6 address.
    # Many services only return one IP version per request.
```

??? info "How dual-stack works"
    Lightddns creates separate IPv4 and IPv6 connections for HTTP datasources. If your network has both, it will discover both addresses. See the `dialStrategy` option in the [ConnectOption reference](../configuration/shared/connect.md) for details.

### Using a Script

If you prefer to get your IP through a custom script:

```yaml
datasources:
  - type: command
    name: my-script
    cmd:
      ipv4: "curl -s https://api.ip.sb/ip"
```

The command's output is scanned for valid IP addresses.

### Reading from a Network Interface

If your machine has a public IP directly on a network interface:

```yaml
datasources:
  - type: netlink
    name: eth0-ip
    ifName: eth0
    allowPrivate: false   # only return public IPs
```

## Multiple Domains

You can manage as many domains as you want — just add more entries under `domains`:

```yaml
domains:
  - domain: home.yourdomain.com
    provider: my-cloudflare
    datasource: my-public-ip

  - domain: vpn.yourdomain.com
    provider: my-cloudflare
    datasource: my-public-ip
    ipv6: true           # this one also gets an AAAA record

  - domain: blog.yourdomain.com
    enabled: false        # temporarily disabled
    provider: my-cloudflare
    datasource: my-public-ip
```

Each domain runs on its own schedule and can use different datasources, providers, or IP families.

## Next Steps

- [How It Works](how-it-works.md) — understand datasources, providers, and group strategies
- [Configuration Reference](../configuration/options.md) — every config option explained
- [HTTP Datasource](../configuration/datasource/http.md) — full HTTP datasource reference
