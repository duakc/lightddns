# Test configurations

These configurations are for manual integration testing before a release.
The release convention is to run the relevant configurations in this folder
successfully before publishing. This convention is intentionally not enforced
by CI.

Set the following environment variables before running a configuration:

- `CLOUDFLARE_TOKEN`
- `ALIYUN_ACCESSKEY_ID` and `ALIYUN_ACCESSKEY_SECRET`
- `TENCENT_SECRET_ID` and `TENCENT_SECRET_KEY`
- `TEST_ZONE_CLOUDFLARE`, `TEST_ZONE_ALIYUN`, and `TEST_ZONE_TENCENT`
- `TEST_HTTP_PROXY` and `TEST_DNS_SERVER` for network option tests
- `TEST_NETLINK_INTERFACE` when running the netlink configuration
- `TEST_NETLINK_IFINDEX` for the numeric netlink/bind-interface scenarios
- `TEST_BIND_ADDRESS4` and `TEST_BIND_ADDRESS6` for bind-address scenarios
- `TEST_FWMARK` for the Linux-only firewall-mark scenario

The command-based provider scenarios read IP addresses from
[`testconfig/testdata/ips`](testdata/ips). Edit this file while Lightddns is running to test
single/multiple address creation, deletion, and replacement. Keep the entries
globally routable because cloud providers reject private or documentation-only
addresses. Credentials, zones, and local network settings are intentionally
not checked in; provide them through the inherited environment or your own
`--env-file`.
The Cloudflare proxy scenario uses [`testconfig/testdata/ips-proxy`](testdata/ips-proxy),
which avoids provider-owned addresses that Cloudflare rejects as proxied targets.
The HTTP proxy tests expect the local sing-box HTTP inbound at the value of
`TEST_HTTP_PROXY` (default: `http://127.0.0.1:9080`). Update
`TEST_NETLINK_INTERFACE` when the test runs on a different host.

The Aliyun and Tencent line scenarios use separate fixtures under
`testconfig/testdata/ips-ct`, `ips-cu`, `ips-cm`, and `ips-default`; each matching
datasource reads only its line's file. Edit one file at a time to observe changes
for that line without changing the others.

The `cloudflare-datasource-command-jq.yaml` scenario reads its JSON fixture from
`testconfig/testdata/ip.json`.

The command streams scenario uses separate `stream-stdout` and `stream-stderr`
fixtures so the captured addresses are distinct.

The `cloudflare-datasource-command-duplicates.yaml` scenario intentionally
reads duplicate IPv4 and IPv6 values from `testdata/ips-duplicates` to verify
that the domain removes duplicates before calling the provider.

For example, from the repository root:

```bash
CLOUDFLARE_TOKEN=... TEST_ZONE_CLOUDFLARE=example.com \
  lightddns -D . check -c testconfig/cloudflare-connect-ipv4.yaml
CLOUDFLARE_TOKEN=... TEST_ZONE_CLOUDFLARE=example.com \
  lightddns -D . run -c testconfig/cloudflare-connect-ipv4.yaml
```

## Coverage map

Each focused file is intended to be run independently. The original combined
files remain useful as smoke tests; the focused files below make one option or
behavior easy to isolate:

- `cloudflare-connect.yaml`: default dial behavior.
- `cloudflare-connect-ipv4.yaml`, `cloudflare-connect-ipv6.yaml`: IPv4-only and
  IPv6-only dialing, with matching domain address-family filters.
- `cloudflare-connect-prefer-ipv4.yaml`, `cloudflare-connect-prefer-ipv6.yaml`:
  the two dual-stack preference strategies.
- `cloudflare-connect-bind-address.yaml` and `cloudflare-connect-bind-address-ipv6.yaml`,
  `cloudflare-connect-bind-interface.yaml`,
  `cloudflare-connect-bind-interface-index.yaml`, `cloudflare-connect-fwmark.yaml`:
  IPv4/IPv6 address, interface-name, interface-index, and firewall-mark binding.
- `cloudflare-http-datasource-connect-ipv4.yaml`,
  `cloudflare-http-datasource-connect-ipv6.yaml`,
  `cloudflare-http-datasource-connect-prefer-ipv4.yaml`, and
  `cloudflare-http-datasource-connect-prefer-ipv6.yaml`: all four dial strategies
  on the HTTP datasource path.
- `cloudflare-http-system-proxy.yaml` and
  `cloudflare-http-proxy-http-only.yaml`: system proxy and single-proxy fallback.
- `cloudflare-dns-system.yaml` and `cloudflare-dns-tls-shorthand.yaml`: DNS
  string forms; `cloudflare-dns.yaml` covers the TLS object form.
- `cloudflare-http.yaml`, `cloudflare-http-system-proxy.yaml`, and
  `cloudflare-http-proxy-http-only.yaml`,
  `cloudflare-http-proxy-https-only.yaml`: provider HTTP proxy modes, including
  either single-value fallback; `cloudflare-http-datasource.yaml` covers HTTP
  datasource GET/POST, headers, JQ/regex matching, DNS, proxy, and debug.
- `cloudflare-datasource-command-jq.yaml` and
  `cloudflare-datasource-command-stdin.yaml`: command JSON extraction and file
  stdin/workdir handling. The file stdin scenario reads `testconfig/testdata/ips`.
  `cloudflare-datasource-command-duplicates.yaml` covers duplicate-address
  removal before provider updates.
  `cloudflare-datasource-command-streams.yaml` covers
  forwarding and parsing both stdout and stderr. `cloudflare-datasources.yaml`
  covers command environment, inline stdin, exit codes, sync, and regex;
  `cloudflare-datasource-command-scalar.yaml` covers scalar `cmd` syntax.
- `cloudflare-datasource-sum.yaml`, `cloudflare-datasource-filter.yaml`, and
  `cloudflare-datasource-failover.yaml`: each datasource group independently.
- `cloudflare-netlink-index.yaml`: numeric interface selection and bogon
  inclusion; `cloudflare-netlink.yaml` covers interface-name selection and
  filtered IPv6 output.
- `cloudflare-domain-defaults.yaml`: provider/datasource auto-selection.
  `cloudflare-domain-services.yaml` covers TTL, IPv4/IPv6 selection, disabled
  domains, and custom service endpoints; `cloudflare-services-defaults.yaml`
  covers service defaults and a disabled service.
- `cloudflare-log-disabled.yaml`: disabled logging.
- `aliyun.yaml` and `tencent.yaml`: independent ct/cu/cm/default line domains
  (using each provider's API values), multiple addresses, and the
  provider-specific default-line initialization behavior. The default-line
  domain is listed first so the other line domains can be initialized safely.
- `cloudflare.yaml` and `cloudflare-proxy.yaml`: basic Cloudflare updates and
  the provider proxy flag.

The bind-address, bind-interface, netlink, and fwmark scenarios depend on the
host OS and network namespace. Adjust their environment values before running
them; they are still valid configuration/schema coverage on every platform.
