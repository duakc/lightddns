# lightddns

`lightddns` 是一个配置优先的 DDNS 服务。

名字里虽然写着 `lightweight`，但项目其实已经不太 `light` 了。现在
它已经长成了一个更完整的工具集，包含多 IP 处理、可复用组件、打包、
服务抽象，以及结构化日志。

当前项目还在往稳定版靠近，所以现在很欢迎社区来直接测试。请优先帮
忙测这些内容：安装包、Provider 更新、`httpProxy` / `httpsProxy`、
自定义 `dns`、服务启动，以及端到端的域名更新。反馈时最好带上包格式、
操作系统、脱敏日志和一小段配置片段。

---

`lightddns` 的核心是一些可以复用的组件。

`datasource` 负责发现 IP，`provider` 负责提交 DNS 变更，`domain` 把两
者绑定起来，`service` 则提供一些可选的运行时能力，比如 Prometheus
指标或者 `ipserver`。

这样的拆分让 transport、DNS、HTTP、datasource、provider 和 service 都
可以复用，而不是被锁死在单一更新路径里。

---

和 [`ddns-go`](https://github.com/jeessy2/ddns-go) 比，关注点不太一样。

`ddns-go` 更成熟，也很适合 Web 配置；

`lightddns` 更强调配置优先和组件化。它支持多 IP、边界更清晰，也更重视打包和运行时复用。当前已经覆
盖 `deb`、`rpm`、`pkg.zst` 和 `Nix`，后面稳定后也会继续往 `AUR` 和
`nixpkgs` 走。

另外它还有结构化日志和 service 抽象，所以运行时更容易排查，也更容易
复用。

---

## 已知限制

现在的 diff 模型支持 Provider 自定义的记录属性。Cloudflare 的 `proxied`
以及 Tencent Cloud、Aliyun 的 `lines` 都可以参与 diff。

网络层还需要更多实测，尤其是连接复用、代理行为、DNS 边界、transport
重试，以及远端代理到底应该收到域名还是本地解析后的 IP。

打包已经做出来了，但还需要更多社区测试，先把它们跑得足够稳定。

文档也还在继续补。配置参考已经有了，但运行时路径还没有全部补齐。

---

## Roadmap

- 补全配置参考。
- 继续完善文档站点。
- 继续扩展 provider 级别的记录属性 diff。
- 继续收紧网络行为、日志和可观测性。

---

## License

`lightddns` 采用 GPL-2.0 许可证，见 [LICENSE](https://github.com/duakc/lightddns/blob/main/LICENSE)。

## Development Note

这个项目的大部分代码是人写的。AI 主要只参与了少量 bug 修复、部分测
试编写，以及重复代码的编写。
