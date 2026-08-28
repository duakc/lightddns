# lightddns

`lightddns` is a configuration-first DDNS service.

The name says lightweight, but the project is not really `light` anymore. It
has grown into a broader toolset with multiple IP handling, reusable
components, packaging, services, and structured logging.

The project is still moving toward a stable release, so community testing is
welcome now. Please try package installs, provider updates, `httpProxy` /
`httpsProxy`, custom `dns`, service startup, and end-to-end domain updates in
real environments. Reports are most useful when they include the package
format, operating system, sanitized logs, and a small configuration snippet.

---

`lightddns` is built from reusable pieces.

A `datasource` discovers IPs. A `provider` applies DNS changes. A `domain`
connects them. A `service` adds optional runtime features such as Prometheus
metrics or the `ipserver` endpoint.

That split keeps transport, DNS, HTTP, datasources, providers, and services
reusable instead of locking them into one update path.

---

Compared with [`ddns-go`](https://github.com/jeessy2/ddns-go), the focus here
is different.

`ddns-go` is mature and very convenient for web-based setup. `lightddns` stays
configuration-first and more composable. It is built to handle multiple IPs,
keep clearer component boundaries, and make packaging and operational reuse
easier. Release artifacts already cover `deb`, `rpm`, `pkg.zst`, and `Nix`,
and the longer-term path is toward `AUR` and `nixpkgs` once the project
settles.

It also has structured logging and service abstractions, so the runtime is
easier to inspect and reuse.

---

## Known Limits

The diff model supports provider-defined record attributes. Cloudflare
`proxied`, and Tencent Cloud and Aliyun `lines`, can participate in diffing.

The network layer still needs more field testing, especially around connection
reuse, proxy behavior, DNS boundaries, transport retries, and how remote
proxies should receive hostnames versus locally resolved IPs.

Packaging exists, but it still needs wider testing from the community before it
can be treated as boring.

The documentation is still growing. The configuration reference is available,
but not every runtime path is documented yet.

---

## Roadmap

- Complete the configuration reference.
- Finish the documentation site.
- Keep extending provider-specific record attributes.
- Keep tightening network behavior, logs, and observability.

---

## License

`lightddns` is licensed under GPL-2.0. See [LICENSE](https://github.com/duakc/lightddns/blob/main/LICENSE).

## Development Note

Most of the code in this project is written by humans. AI assistance is used
for a limited part of the work, mainly bug fixes, some test writing, and
repetitive code.
