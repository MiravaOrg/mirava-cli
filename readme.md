# Mirava CLI

A command-line tool for testing package registry mirrors. Check mirror health, verify package availability, and measure download speeds across popular ecosystems.

Built on mirava-core

## Supported ecosystems

| Ecosystem | Commands |
|-----------|----------|
| npm | `status`, `package`, `speed` |
| PyPI | `status`, `package`, `speed` |
| Go | `status`, `package`, `speed` |
| Cargo (Rust) | `status`, `package`, `speed` |
| Composer (PHP) | `status`, `package`, `speed` |
| Docker | `status`, `package`, `speed` |
| apt (generic) | `status`, `package`, `speed` |
| ubuntu / debian | `status`, `package`, `speed` |
| pacman (Arch) | `status`, `package`, `speed` |

## Installation

### Linux

Download the archive for your architecture from the [latest GitHub release](https://github.com/miravaOrg/mirava-cli/releases/latest), or install straight from the terminal:

```bash
# amd64 (most Linux distros)
curl -LO https://github.com/miravaOrg/mirava-cli/releases/latest/download/mirava-cli-linux-amd64.tar.gz
tar -xzf mirava-cli-linux-amd64.tar.gz
sudo mv mirava-cli /usr/local/bin/

# arm64 (e.g. Raspberry Pi, ARM servers)
curl -LO https://github.com/miravaOrg/mirava-cli/releases/latest/download/mirava-cli-linux-arm64.tar.gz
tar -xzf mirava-cli-linux-arm64.tar.gz
sudo mv mirava-cli /usr/local/bin/
```

### macOS

```bash
# Apple Silicon (M1/M2/M3/M4)
curl -LO https://github.com/miravaOrg/mirava-cli/releases/latest/download/mirava-cli-darwin-arm64.tar.gz
tar -xzf mirava-cli-darwin-arm64.tar.gz
sudo mv mirava-cli /usr/local/bin/

# Intel Mac
curl -LO https://github.com/miravaOrg/mirava-cli/releases/latest/download/mirava-cli-darwin-amd64.tar.gz
tar -xzf mirava-cli-darwin-amd64.tar.gz
sudo mv mirava-cli /usr/local/bin/
```

On first run, macOS Gatekeeper may block the unsigned binary. Allow it under
**System Settings → Privacy & Security** ("Allow Anyway"), or clear the
quarantine flag directly:

```bash
xattr -d com.apple.quarantine /usr/local/bin/mirava-cli
```

### Windows

Download the zip for your architecture from the
[latest GitHub release](https://github.com/miravaOrg/mirava-cli/releases/latest)
(`mirava-cli-windows-amd64.zip` for x64, `mirava-cli-windows-arm64.zip` for
ARM), then:

```powershell
# In PowerShell, from your Downloads folder
Expand-Archive mirava-cli-windows-amd64.zip -DestinationPath .
New-Item -ItemType Directory -Force "$env:LOCALAPPDATA\Programs\mirava" | Out-Null
Move-Item mirava-cli.exe "$env:LOCALAPPDATA\Programs\mirava\"
```

Add it to your PATH (persistent, current user):

```powershell
[Environment]::SetEnvironmentVariable("Path", [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:LOCALAPPDATA\Programs\mirava", "User")
```

Open a new terminal and run `mirava-cli --help` to confirm.

### Verify the download (all platforms)

Each release ships a `checksums.txt` with SHA-256 hashes. Verify before installing:

```bash
# Linux / macOS
shasum -a 256 -c checksums.txt        # macOS
sha256sum -c checksums.txt            # Linux
```

```powershell
# Windows (PowerShell)
Get-FileHash mirava-cli-windows-amd64.zip -Algorithm SHA256
```

Compare the output with the matching line in `checksums.txt`.

### Build from source

Requires Go 1.26.3 or later.

```bash
git clone https://github.com/miravaOrg/mirava-cli.git
cd mirava-cli
go build -o mirava-cli .
```

## Usage

Each ecosystem follows the same command pattern:

```bash
mirava-cli <ecosystem> <command> [args] [flags]
```

### Check mirror status

Verify that a mirror responds correctly.

```bash
mirava-cli npm status https://registry.npmmirror.com
mirava-cli go status https://proxy.golang.org
mirava-cli apt status https://mirrors.aliyun.com/ubuntu/
```

### apt package lookup options

The `apt`, `ubuntu` and `debian` commands share the same flags. When
release/component are omitted, the backend discovers suites from the
mirror's own `InRelease`/`Release` metadata; arch defaults to `amd64`.

```bash
mirava-cli apt package https://mirrors.aliyun.com/ubuntu/ curl -a amd64 -c main -r noble
mirava-cli ubuntu package https://mirrors.aliyun.com/ubuntu/ curl
mirava-cli debian package https://mirrors.aliyun.com/debian/ curl
```

| Flag | Description | Default |
|------|-------------|---------|
| `-a`, `--arch` | Package architecture (e.g. `amd64`, `arm64`) | `amd64` |
| `-c`, `--component` | Repository component (e.g. `main`, `universe`) | discovered |
| `-r`, `--release` | Release or codename (e.g. `noble`, `bookworm`) | discovered |

### Check package availability

Confirm a specific package exists on the mirror.

```bash
mirava-cli npm package https://registry.npmmirror.com lodash
mirava-cli pypi package https://pypi.org requests
mirava-cli cargo package https://crates.io serde
```

### Measure download speed

Test how fast a mirror serves package downloads.

```bash
mirava-cli npm speed https://registry.npmmirror.com -p lodash -t 15
mirava-cli go speed https://proxy.golang.org -t 10
```

| Flag | Description | Default |
|------|-------------|---------|
| `-p`, `--package` | Package name for speed tests | — |
| `-t`, `--timeout` | Download timeout in seconds | `10` |

## Examples

```bash
# NPM mirror health check
mirava-cli npm status https://registry.npmmirror.com

# PyPI package lookup
mirava-cli pypi package https://pypi.tuna.tsinghua.edu.cn/simple requests

# Docker registry speed test
mirava-cli docker speed https://registry-1.docker.io -p alpine -t 20

# Pacman mirror status
mirava-cli pacman status https://mirrors.kernel.org/archlinux/
```

## Releasing

Releases are automated via GitHub Actions.

### Trigger manually (workflow dispatch)

1. Open **Actions** → **Release** → **Run workflow**
2. Enter a semver tag (e.g. `v1.0.0`)
3. Run the workflow — binaries for all platforms are built and published to a GitHub release

### Trigger via tag push

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Release artifacts

Each release includes binaries for:

- `linux` — `amd64`, `arm64`
- `darwin` — `amd64`, `arm64`
- `windows` — `amd64`, `arm64`

Plus a `checksums.txt` with SHA-256 hashes.

## Development

```bash
# Run without installing
go run .

# Build locally
go build -o mirava-cli .

# Cross-compile (example)
GOOS=linux GOARCH=arm64 go build -o mirava-cli-linux-arm64 .
```

## License

See the repository license file for details.
