# Manual

Lightddns is built from five composable pieces that together form a complete DDNS pipeline:

```
┌────────────┐   ┌──────────┐   ┌──────────┐
│ Datasource ├──▶│  Domain  ├──▶│ Provider │
└────────────┘   └──────────┘   └──────────┘
       discovers IP    schedules updates    pushes to DNS

                  ┌──────────┐   ┌──────────┐
                  │   Log    │   │ Services │
                  └──────────┘   └──────────┘
                  global logger   prometheus / ipserver
```

| Section | Role | Pages |
|---|---|---|
| `log` | Global logger output and level | [Log](../configuration/log.md) |
| `datasources` | How to learn the current public IP — HTTP API, network interface, shell command, or a group of any of these | [Datasources](../configuration/options.md#datasources) |
| `providers` | How to talk to the DNS service — Cloudflare, Aliyun, Tencent Cloud | [Providers](../configuration/options.md#providers) |
| `domains` | The bind: which provider updates which name from which datasource, on what schedule | [Domain](../configuration/domain.md) |
| `services` | Optional background HTTP services — Prometheus metrics exporter, IP echo server | [Services](../configuration/options.md#services) |

## The update loop

For every enabled domain, Lightddns:

1. Asks its datasource for the current IP (`interval` apart).
2. Diffs the IP against the records currently in DNS.
3. Calls the provider to create / update / delete records only when the diff is non-empty.

Only step 3 ever talks to your DNS API, so a stable IP costs zero API calls per cycle.

## Config files are Go templates

The YAML file is rendered as a [Go template](https://pkg.go.dev/text/template) before parsing. Use `{{ .Env.NAME }}` to pull values from the OS environment or a `.env` file in the working directory:

```yaml
providers:
  - type: cloudflare
    token: "{{ .Env.CLOUDFLARE_TOKEN }}"
```

This keeps credentials out of the committed config.

## Running

```bash
# continuous
lightddns run -c example.yaml

# one-shot (single update pass, then exit)
lightddns run -c example.yaml --once
```

## Examples

Each example below is a self-contained config. Drop one into `lightddns.yaml`, set the referenced env vars, and run.

- [Cloudflare](cloudflare.md)
- [Aliyun](aliyun.md)
- [Tencent Cloud](tencentcloud.md)
- [Group datasources (sum / failover)](groups.md)
- [Services (Prometheus / IP echo)](services.md)