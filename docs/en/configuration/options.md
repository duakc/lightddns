# Configuration File

Structure overview of the Lightddns configuration file. See sub-pages for detailed field references.

```yaml
# required
datasources:
  - type: http
    name: data-http
    # ... HTTP datasource config
providers:
  - type: cloudflare
    name: prov-cf
    # ... Cloudflare provider config
domains:
  - domain: example.com
    provider: prov-cf
    datasource: data-http

log:
  level: info
```

## `datasources`

A list of datasources. Each datasource discovers the current public IP address of the host.

See [HTTP Datasource](datasource/http.md).

## `providers`

A list of Service Providers. Each provider pushes IP updates to the corresponding DNS service.

## `domains`

A list of domain records. Each entry binds a domain, provider, and datasource together for automatic DDNS updates.

See [Domain](domain.md).

## `log`

Global logging configuration. See [Log](log.md).
