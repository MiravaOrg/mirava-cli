# Mirava CLI

A command-line tool for testing package registry mirrors. Check mirror health, verify package availability, and measure download speeds across popular ecosystems.

Built on [mirava-core](https://github.com/MiravaOrg/mirava-core).

## Supported ecosystems

| Ecosystem | Commands |
|-----------|----------|
| npm | `status`, `package`, `speed` |
| PyPI | `status`, `package`, `speed` |
| Go | `status`, `package`, `speed` |
| Cargo (Rust) | `status`, `package`, `speed` |
| Composer (PHP) | `status`, `package`, `speed` |
| Docker | `status`, `package`, `speed` |
| apt (Debian/Ubuntu) | `status`, `package`, `speed` |
| pacman (Arch) | `status`, `package`, `speed` |

## Installation

### Pre-built binaries

Download the archive for your platform from the [latest GitHub release](https://github.com/miravaOrg/mirava-cli/releases/latest).

```bash
# Example: Linux amd64
curl -LO https://github.com/miravaOrg/mirava-cli/releases/latest/download/mirava-cli-linux-amd64.tar.gz
tar -xzf mirava-cli-linux-amd64.tar.gz
sudo mv mirava-cli /usr/local/bin/
```

Verify the checksum against `checksums.txt` in the release assets.

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

Use `-v` / `--verbose` for detailed output (where supported).

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
| `-v`, `--verbose` | Verbose output | `false` |
| `-p`, `--package` | Package name for speed tests | — |
| `-t`, `--timeout` | Download timeout in seconds | `10` |

## Examples

```bash
# NPM mirror health check
mirava-cli npm status https://registry.npmmirror.com -v

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
