# lightddns

`lightddns` is a lightweight, configuration-first DDNS service.

The project is still moving toward a stable release, so community testing is
very welcome right now. Please try the packages, providers, `httpProxy`,
`httpsProxy`, custom `dns`, and service startup behavior in real environments.

Useful reports usually include the package format, operating system, sanitized
logs, and a small configuration snippet.

---

`lightddns` is built around a few reusable ideas.

A `datasource` discovers IP addresses. A `provider` applies DNS record updates.
A `domain` connects a datasource with a provider. A `service` exposes optional
runtime features that are useful around DDNS, such as `prometheus` metrics or an
`ipserver` endpoint.

This layout is meant to keep the project flexible. The same `connect`, `dns`,
`http`, datasource, provider, and service pieces can be reused in different
places instead of being locked inside one update path.

---

The project is different from
[`ddns-go`](https://github.com/jeessy2/ddns-go) in a few important ways.

`ddns-go` is mature, popular, and very convenient for web-based configuration.
`lightddns` is not trying to be a drop-in replacement for it. The main goal here
is to make DDNS workflows easier to compose, test, package, and run from a
configuration file.

`lightddns` can work with multiple IP addresses, not only one simple result. It
also gives more structure to datasources and providers, so IP discovery,
filtering, failover, and DNS record updates can be combined more freely.

The project also puts more attention on packaging. Release work already covers
formats such as `deb`, `rpm`, `pkg.zst`, and `Nix`. When the project is stable
enough, packages can move toward community repositories such as `AUR` and
`nixpkgs`.

Another difference is runtime visibility. `lightddns` has structured logging
and service abstractions, so debugging datasource, provider, DNS, proxy, and
transport behavior should be clearer over time.

---

## Current Limits

The biggest known limitation is the diff model.

Right now, diffing mainly works on IP addresses. Provider-specific record state
is not fully represented yet. For example, Cloudflare's `proxied` status cannot
be reliably diffed and updated when the IP itself has not changed.

The same limitation affects planned Tencent Cloud and Aliyun `line` support.
Those features need the diff model to understand provider-specific record
attributes, not only record IPs.

---

The network layer also needs more refinement.

Connection reuse, `dns` boundaries, HTTP proxy behavior, transport behavior, and
retry rules all need more real-world testing. One important rule is that when
`httpProxy` or `httpsProxy` is enabled, target domains should be sent to the
remote proxy. Local `dns` configuration is only expected to affect the address
being dialed locally, such as the proxy host itself.

---

Packaging needs testing too.

The release outputs exist, but they have not been tested by many users yet.
Please test installation, upgrade, service startup, service restart, logs, and
uninstall behavior for `deb`, `rpm`, `pkg.zst`, and `Nix`.

---

The documentation is not complete yet.

The goal is to provide a full reference for every configuration option and
runtime behavior. Until that is finished, some details may still require reading
examples or source code.

---

## Roadmap

- [x] Add `prometheus` metrics.
- [x] Add Tencent Cloud and Aliyun providers.
- [x] Add `deb`, `rpm`, `Nix`, `pkg.zst`, and related release packaging.
- [ ] Complete the configuration reference.
- [ ] Add a web configuration generator in the documentation site.
- [ ] Extend diffing to provider-specific record attributes.
- [ ] Add reliable Tencent Cloud and Aliyun `line` support.
- [ ] Continue improving network behavior, logs, and observability.

## Considering

Webhook integrations such as `Slack` or `Telegram` are still being considered.
External alerting systems can already consume `prometheus` metrics, so native
webhooks should only be added if they expose useful runtime state without making
the core service unnecessarily complex.

## License

`lightddns` is licensed under the GNU General Public License v2.0. See
[LICENSE](LICENSE) for details.

## Development Notice

Most code in this project is written and designed by humans. AI assistance is
used for a limited part of the work, mainly some bug fixes, some test writing,
and repetitive code.
