# 安装

Lightddns 以单个二进制文件分发——无需安装程序、无需依赖，下载即可运行。

## 下载预编译版本

前往 [GitHub Releases](https://github.com/duakc/lightddns/releases) 页面，下载对应你平台的压缩包：

| 平台 | 文件名 |
|---|---|
| Linux (x86_64) | `lightddns-linux-amd64.tar.gz` |
| Linux (ARM64) | `lightddns-linux-arm64.tar.gz` |
| macOS (Apple Silicon) | `lightddns-darwin-arm64.tar.gz` |
| macOS (Intel) | `lightddns-darwin-amd64.tar.gz` |
| Windows (x86_64) | `lightddns-windows-amd64.zip` |

解压后将 `lightddns` 放入 `PATH` 中的任意目录——Linux 和 macOS 上通常放在 `/usr/local/bin`。

=== "Linux / macOS"

    ```bash
    # 下载（将 URL 替换为实际下载地址）
    curl -LO https://github.com/duakc/lightddns/releases/latest/download/lightddns-linux-amd64.tar.gz

    # 解压
    tar xzf lightddns-linux-amd64.tar.gz

    # 安装
    sudo mv lightddns /usr/local/bin/
    ```

=== "Windows"

    下载 `.zip` 文件，解压后将 `lightddns.exe` 放入 `PATH` 中的任意文件夹。

## 验证安装

```bash
lightddns version
```

如果输出了版本号和 git commit，说明安装成功。

## 从源码构建

如果你安装了 Go 1.26+：

```bash
git clone https://github.com/duakc/lightddns.git
cd lightddns
make build
```

编译好的二进制文件在 `build/lightddns`。

## 配置后台服务

在 Linux 上，可以将 Lightddns 配置为 systemd 服务，实现开机自启和后台运行。

创建 `/etc/systemd/system/lightddns.service`：

```ini
[Unit]
Description=Lightddns 动态 DNS 更新服务
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/lightddns run -c /etc/lightddns/lightddns.yaml
Restart=always
RestartSec=30
User=nobody

[Install]
WantedBy=multi-user.target
```

然后启用并启动：

```bash
sudo mkdir -p /etc/lightddns
sudo cp lightddns.yaml /etc/lightddns/
sudo systemctl daemon-reload
sudo systemctl enable --now lightddns
```

查看运行状态：

```bash
sudo systemctl status lightddns
```

## 下一步

安装完成后，按照[快速开始](getting-started.md)指南完成首个域名的配置。
