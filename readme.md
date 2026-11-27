# mirava

A command-line tool for testing package registry mirrors. Check mirror health, verify package availability, and measure download speeds across popular ecosystems.

Built on mirava-core

## Supported ecosystems

| Ecosystem | Commands |
|-----------|----------|
| npm | `status`, `package`, `speed`, `set`, `get`, `reset` |
| PyPI | `status`, `package`, `speed`, `set`, `get`, `reset` |
| Go | `status`, `package`, `speed`, `set`, `get`, `reset` |
| Cargo (Rust) | `status`, `package`, `speed`, `set`, `get`, `reset` |
| Composer (PHP) | `status`, `package`, `speed`, `set`, `get`, `reset` |
| Docker | `status`, `package`, `speed`, `set`, `get`, `reset` |
| apt (generic) | `status`, `package`, `speed` |
| ubuntu / debian | `status`, `package`, `speed`, `set`, `get`, `reset` |
| pacman (Arch) | `status`, `package`, `speed`, `set`, `get`, `reset` |

## Installation

### Linux

Download the archive for your architecture from the [latest GitHub release](https://github.com/miravaOrg/mirava/releases/latest), or install straight from the terminal:

```bash
# amd64 (most Linux distros)
curl -LO https://github.com/miravaOrg/mirava/releases/latest/download/mirava-linux-amd64.tar.gz
tar -xzf mirava-linux-amd64.tar.gz
sudo mv mirava /usr/local/bin/

# arm64 (e.g. Raspberry Pi, ARM servers)
curl -LO https://github.com/miravaOrg/mirava/releases/latest/download/mirava-linux-arm64.tar.gz
tar -xzf mirava-linux-arm64.tar.gz
sudo mv mirava /usr/local/bin/
```

### macOS

```bash
# Apple Silicon (M1/M2/M3/M4)
curl -LO https://github.com/miravaOrg/mirava/releases/latest/download/mirava-darwin-arm64.tar.gz
tar -xzf mirava-darwin-arm64.tar.gz
sudo mv mirava /usr/local/bin/

# Intel Mac
curl -LO https://github.com/miravaOrg/mirava/releases/latest/download/mirava-darwin-amd64.tar.gz
tar -xzf mirava-darwin-amd64.tar.gz
sudo mv mirava /usr/local/bin/
```

On first run, macOS Gatekeeper may block the unsigned binary. Allow it under
**System Settings → Privacy & Security** ("Allow Anyway"), or clear the
quarantine flag directly:

```bash
xattr -d com.apple.quarantine /usr/local/bin/mirava
```

### Windows

Download the zip for your architecture from the
[latest GitHub release](https://github.com/miravaOrg/mirava/releases/latest)
(`mirava-windows-amd64.zip` for x64, `mirava-windows-arm64.zip` for
ARM), then:

```powershell
# In PowerShell, from your Downloads folder
Expand-Archive mirava-windows-amd64.zip -DestinationPath .
New-Item -ItemType Directory -Force "$env:LOCALAPPDATA\Programs\mirava" | Out-Null
Move-Item mirava.exe "$env:LOCALAPPDATA\Programs\mirava\"
```

Add it to your PATH (persistent, current user):

```powershell
[Environment]::SetEnvironmentVariable("Path", [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:LOCALAPPDATA\Programs\mirava", "User")
```

Open a new terminal and run `mirava --help` to confirm.

### Verify the download (all platforms)

Each release ships a `checksums.txt` with SHA-256 hashes. Verify before installing:

```bash
# Linux / macOS
shasum -a 256 -c checksums.txt        # macOS
sha256sum -c checksums.txt            # Linux
```

```powershell
# Windows (PowerShell)
Get-FileHash mirava-windows-amd64.zip -Algorithm SHA256
```

Compare the output with the matching line in `checksums.txt`.

### Build from source

Requires Go 1.26.3 or later.

```bash
git clone https://github.com/miravaOrg/mirava.git
cd mirava
go build -o mirava .
```

## Usage

Each ecosystem follows the same command pattern:

```bash
mirava <ecosystem> <command> [args] [flags]
```

### Check mirror status

Verify that a mirror responds correctly.

```bash
mirava npm status https://registry.npmmirror.com
mirava go status https://proxy.golang.org
mirava apt status https://mirrors.aliyun.com/ubuntu/
```

### apt package lookup options

The `apt`, `ubuntu` and `debian` commands share the same flags. When
release/component are omitted, the backend discovers suites from the
mirror's own `InRelease`/`Release` metadata; arch defaults to `amd64`.

```bash
mirava apt package https://mirrors.aliyun.com/ubuntu/ curl -a amd64 -c main -r noble
mirava ubuntu package https://mirrors.aliyun.com/ubuntu/ curl
mirava debian package https://mirrors.aliyun.com/debian/ curl
```

| Flag | Description | Default |
|------|-------------|---------|
| `-a`, `--arch` | Package architecture (e.g. `amd64`, `arm64`) | `amd64` |
| `-c`, `--component` | Repository component (e.g. `main`, `universe`) | discovered |
| `-r`, `--release` | Release or codename (e.g. `noble`, `bookworm`) | discovered |

### Check package availability

Confirm a specific package exists on the mirror.

```bash
mirava npm package https://registry.npmmirror.com lodash
mirava pypi package https://pypi.org requests
mirava cargo package https://crates.io serde
```

### Measure download speed

Test how fast a mirror serves package downloads.

```bash
mirava npm speed https://registry.npmmirror.com -p lodash -t 15
mirava go speed https://proxy.golang.org -t 10
```

| Flag | Description | Default |
|------|-------------|---------|
| `-p`, `--package` | Package name for speed tests | — |
| `-t`, `--timeout` | Download timeout in seconds | `10` |

## Switching your system to a mirror

Beyond testing, mirava can switch your system's package tooling over to a
mirror with a single command — and switch it back.

```bash
mirava set <tech> <mirror-url>     # switch to a mirror
mirava get <tech>                  # show what is currently set
mirava reset <tech>                # restore the default upstream
```

Every `set` verifies the mirror is reachable first and aborts on failure;
pass `--force` to skip that check. `reset` restores the default upstream
registry (or the original config from the backup `set` made).

### set

```bash
mirava set npm https://registry.npmmirror.com        # npm config set registry
mirava set pypi https://pypi.tuna.tsinghua.edu.cn/simple   # pip config set global.index-url
mirava set go https://goproxy.cn                     # go env -w GOPROXY
mirava set cargo https://rustmirror.example.com      # ~/.cargo/config.toml source replacement
mirava set composer https://packagist.example.com    # composer global repo
mirava set docker https://docker.mirrors.example.com # daemon.json registry-mirrors
mirava set ubuntu https://mirrors.aliyun.com         # apt sources (Ubuntu)
mirava set debian https://mirrors.aliyun.com         # apt sources (Debian)
mirava set pacman https://mirrors.kernel.org/archlinux/  # /etc/pacman.d/mirrorlist
```

Notes:

- `set ubuntu` / `set debian` refuse to run on any other distro, and write
  the format your release actually uses: classic `/etc/apt/sources.list` on
  older releases, deb822 `sources.list.d/mirava.sources` on Debian 13+ and
  Ubuntu 24.04+. The original files are kept as `*.mirava-bak`.
- `set docker` adds the mirror to `registry-mirrors` in
  `/etc/docker/daemon.json` (preserving other keys) and reloads the daemon.
- `set pacman` replaces `/etc/pacman.d/mirrorlist` with your mirror and runs
  `pacman -Sy`.
- apt, pacman and docker changes require root; run with `sudo`.

### get

```bash
mirava get npm       # Registry: https://registry.npmmirror.com
mirava get pypi      # Index URL: ...
mirava get go        # GOPROXY: ...
mirava get cargo     # prints ~/.cargo/config.toml
mirava get docker    # lists configured registry mirrors
mirava get ubuntu    # prints apt sources (sources.list + sources.list.d/)
mirava get pacman    # lists uncommented Server lines
mirava get composer  # prints global repositories
```

### reset

```bash
mirava reset npm       # back to registry.npmjs.org
mirava reset pypi      # back to pypi.org
mirava reset go        # back to https://proxy.golang.org,direct
mirava reset cargo     # removes the mirror source from ~/.cargo/config.toml
mirava reset composer  # removes the mirror repo, re-enables packagist
mirava reset docker    # removes registry-mirrors from daemon.json
mirava reset ubuntu    # restores apt sources from the mirava backup
mirava reset debian    # restores apt sources from the mirava backup
mirava reset pacman    # restores the mirrorlist from the mirava backup
```

## Examples

```bash
# NPM mirror health check
mirava npm status https://registry.npmmirror.com

# PyPI package lookup
mirava pypi package https://pypi.tuna.tsinghua.edu.cn/simple requests

# Docker registry speed test
mirava docker speed https://registry-1.docker.io -p alpine -t 20

# Pacman mirror status
mirava pacman status https://mirrors.kernel.org/archlinux/

# Switch npm to a mirror, check it, then go back
mirava set npm https://registry.npmmirror.com
mirava get npm
mirava reset npm
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
go build -o mirava .

# Cross-compile (example)
GOOS=linux GOARCH=arm64 go build -o mirava-linux-arm64 .
```

## License

See the repository license file for details.
