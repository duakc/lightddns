# Installation

Lightddns is distributed as a single binary — no installer, no dependencies, just download and run.

## Download Pre-built Binary

Go to the [GitHub Releases](https://github.com/duakc/lightddns/releases) page and download the archive for your platform:

| Platform | File |
|---|---|
| Linux (x86_64) | `lightddns-linux-amd64.tar.gz` |
| Linux (ARM64) | `lightddns-linux-arm64.tar.gz` |
| macOS (Apple Silicon) | `lightddns-darwin-arm64.tar.gz` |
| macOS (Intel) | `lightddns-darwin-amd64.tar.gz` |
| Windows (x86_64) | `lightddns-windows-amd64.zip` |

Extract the archive and place the `lightddns` binary anywhere in your `PATH` — `/usr/local/bin` is a common choice on Linux and macOS.

=== "Linux / macOS"

    ```bash
    # Download (replace URL with the actual release link)
    curl -LO https://github.com/duakc/lightddns/releases/latest/download/lightddns-linux-amd64.tar.gz

    # Extract
    tar xzf lightddns-linux-amd64.tar.gz

    # Install
    sudo mv lightddns /usr/local/bin/
    ```

=== "Windows"

    Download the `.zip` file, extract it, and place `lightddns.exe` in a folder that's in your `PATH`.

## Verify Installation

```bash
lightddns version
```

This prints the version number and git commit. If you see output, you're ready to go.

## Build from Source

If you have Go 1.26+ installed:

```bash
git clone https://github.com/duakc/lightddns.git
cd lightddns
make build
```

The binary will be at `build/lightddns`.

## Running as a Background Service

On Linux, you can run Lightddns as a systemd service so it starts automatically and runs in the background.

Create `/etc/systemd/system/lightddns.service`:

```ini
[Unit]
Description=Lightddns Dynamic DNS Updater
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

Then enable and start:

```bash
sudo mkdir -p /etc/lightddns
sudo cp lightddns.yaml /etc/lightddns/
sudo systemctl daemon-reload
sudo systemctl enable --now lightddns
```

Check status:

```bash
sudo systemctl status lightddns
```

## Next Steps

Now that Lightddns is installed, follow the [Getting Started](getting-started.md) guide to set up your first domain.
