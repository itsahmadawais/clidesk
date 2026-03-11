# CLIDesk — Distribution Guide

This document covers every method for getting CLIDesk into users' hands, from the simplest `go install` to proper package manager listings.

---

## 1. `go install` (Works Today)

The fastest path. Any developer with Go installed can run:

```bash
go install github.com/itsahmadawais/clidesk@latest
```

The binary lands in `$GOPATH/bin` (typically `~/go/bin` or `C:\Users\<user>\go\bin`), which is usually already in `$PATH`.

**Requirements:** Go 1.21+  
**Platforms:** Windows, macOS, Linux (any architecture Go supports)

---

## 2. Pre-built Binaries via GitHub Releases (Recommended for General Users)

Non-Go users need pre-built binaries. Set this up once and all other methods (Homebrew, Scoop, Winget) can point to these releases.

### Step 1 — Cross-compile locally

```bash
# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/clidesk-windows-amd64.exe .

# macOS (Intel)
GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/clidesk-darwin-amd64 .

# macOS (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/clidesk-darwin-arm64 .

# Linux (64-bit)
GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/clidesk-linux-amd64 .

# Linux (ARM64 — Raspberry Pi, etc.)
GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o dist/clidesk-linux-arm64 .
```

> `-ldflags="-s -w"` strips debug symbols, reducing binary size by ~30%.

### Step 2 — Create checksums

```bash
cd dist
sha256sum * > checksums.txt
```

### Step 3 — Create a GitHub Release

```bash
gh release create v1.0.0 dist/* \
  --repo itsahmadawais/clidesk \
  --title "CLIDesk v1.0.0" \
  --notes "Initial release"
```

---

## 3. Automate with GitHub Actions

The workflow files are already included in the repo at `.github/workflows/`.

**`ci.yml`** — runs on every push and pull request; builds on Windows, macOS, and Linux in parallel to catch cross-platform issues early.

**`release.yml`** — triggers when you push a version tag; cross-compiles for all 5 targets, creates zip/tar.gz archives, generates a `checksums.txt`, and publishes a GitHub Release automatically.

To trigger a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

That's it. The workflow produces:

```
clidesk-windows-amd64.zip
clidesk-windows-arm64.zip
clidesk-darwin-amd64.tar.gz
clidesk-darwin-arm64.tar.gz
clidesk-linux-amd64.tar.gz
clidesk-linux-arm64.tar.gz
checksums.txt
```

Pre-release versions (e.g. `v1.0.0-beta.1`) are automatically marked as pre-release on GitHub.

### The workflow in detail

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build all platforms
        run: |
          mkdir -p dist
          GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/clidesk-windows-amd64.exe .
          GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/clidesk-darwin-amd64  .
          GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/clidesk-darwin-arm64  .
          GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/clidesk-linux-amd64   .
          GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o dist/clidesk-linux-arm64   .
          cd dist && sha256sum * > checksums.txt

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
```

Trigger a release by pushing a tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Using GoReleaser (alternative, more powerful)

[GoReleaser](https://goreleaser.com/) automates cross-compilation, archives, checksums, changelogs, and package manager manifests in one tool.

```bash
brew install goreleaser   # or go install github.com/goreleaser/goreleaser/v2@latest
goreleaser init           # generates .goreleaser.yaml
goreleaser release --clean
```

---

## 4. Homebrew (macOS & Linux)

Homebrew is the most common package manager on macOS. Users can install with:

```bash
brew install itsahmadawais/tap/clidesk
```

### Create a Homebrew Tap

1. Create a new GitHub repo named `homebrew-tap`
2. Add a formula file at `Formula/clidesk.rb`:

```ruby
class Clidesk < Formula
  desc "Desktop-style file explorer for the terminal"
  homepage "https://github.com/itsahmadawais/clidesk"
  version "1.0.0"

  on_macos do
    on_intel do
      url "https://github.com/itsahmadawais/clidesk/releases/download/v#{version}/clidesk-darwin-amd64"
      sha256 "PASTE_SHA256_HERE"
    end
    on_arm do
      url "https://github.com/itsahmadawais/clidesk/releases/download/v#{version}/clidesk-darwin-arm64"
      sha256 "PASTE_SHA256_HERE"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/itsahmadawais/clidesk/releases/download/v#{version}/clidesk-linux-amd64"
      sha256 "PASTE_SHA256_HERE"
    end
    on_arm do
      url "https://github.com/itsahmadawais/clidesk/releases/download/v#{version}/clidesk-linux-arm64"
      sha256 "PASTE_SHA256_HERE"
    end
  end

  def install
    bin.install Dir["clidesk*"].first => "clidesk"
  end

  test do
    assert_match "CLIDesk", shell_output("#{bin}/clidesk --help 2>&1", 2)
  end
end
```

3. Users install via:
```bash
brew tap itsahmadawais/tap
brew install clidesk
```

> GoReleaser can auto-generate and push this formula on every release.

---

## 5. Scoop (Windows)

[Scoop](https://scoop.sh/) is a popular Windows package manager for developers.

```powershell
scoop bucket add itsahmadawais https://github.com/itsahmadawais/scoop-bucket
scoop install clidesk
```

### Create a Scoop Bucket

1. Create a GitHub repo named `scoop-bucket`
2. Add `clidesk.json`:

```json
{
  "version": "1.0.0",
  "description": "Desktop-style file explorer for the terminal",
  "homepage": "https://github.com/itsahmadawais/clidesk",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "https://github.com/itsahmadawais/clidesk/releases/download/v1.0.0/clidesk-windows-amd64.exe",
      "hash": "PASTE_SHA256_HERE",
      "bin": "clidesk-windows-amd64.exe"
    }
  },
  "checkver": {
    "github": "https://github.com/itsahmadawais/clidesk"
  },
  "autoupdate": {
    "architecture": {
      "64bit": {
        "url": "https://github.com/itsahmadawais/clidesk/releases/download/v$version/clidesk-windows-amd64.exe"
      }
    }
  }
}
```

> GoReleaser can auto-generate and push this manifest.

---

## 6. Winget (Windows — Microsoft Store ecosystem)

[winget](https://learn.microsoft.com/en-us/windows/package-manager/) is Windows' built-in package manager (Windows 10 1709+).

```powershell
winget install itsahmadawais.clidesk
```

### Submit to winget-pkgs

1. Fork [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs)
2. Add a manifest under `manifests/y/itsahmadawais/clidesk/1.0.0/`:

`itsahmadawais.clidesk.installer.yaml`:
```yaml
PackageIdentifier: itsahmadawais.clidesk
PackageVersion: 1.0.0
InstallerType: portable
Installers:
  - Architecture: x64
    InstallerUrl: https://github.com/itsahmadawais/clidesk/releases/download/v1.0.0/clidesk-windows-amd64.exe
    InstallerSha256: PASTE_SHA256_HERE
ManifestType: installer
ManifestVersion: 1.6.0
```

`itsahmadawais.clidesk.locale.en-US.yaml`:
```yaml
PackageIdentifier: itsahmadawais.clidesk
PackageVersion: 1.0.0
PackageLocale: en-US
Publisher: itsahmadawais
PackageName: CLIDesk
ShortDescription: Desktop-style file explorer for the terminal
License: MIT
PackageUrl: https://github.com/itsahmadawais/clidesk
ManifestType: locale
ManifestVersion: 1.6.0
```

3. Submit a PR to `microsoft/winget-pkgs`. Microsoft will review and merge it.

---

## 7. Snap (Linux — Ubuntu/Snapcraft)

```bash
sudo snap install clidesk
```

### Create a Snap

1. Install snapcraft: `sudo snap install snapcraft --classic`
2. Create `snap/snapcraft.yaml`:

```yaml
name: clidesk
version: '1.0.0'
summary: Desktop-style file explorer for the terminal
description: |
  CLIDesk renders your filesystem as an icon grid inside the terminal,
  with keyboard navigation, a built-in command runner, and git status indicators.

grade: stable
confinement: strict
base: core22

parts:
  clidesk:
    plugin: go
    source: .
    build-snaps:
      - go/1.21/stable

apps:
  clidesk:
    command: bin/clidesk
    plugs:
      - home
      - removable-media
```

3. Build and publish:
```bash
snapcraft
snapcraft upload clidesk_1.0.0_amd64.snap --release=stable
```

---

## 8. AUR (Arch Linux)

Arch users can install from the AUR:

```bash
yay -S clidesk-bin
```

### Create an AUR Package

1. Create an account at https://aur.archlinux.org
2. Create a new package `clidesk-bin`
3. Write `PKGBUILD`:

```bash
pkgname=clidesk-bin
pkgver=1.0.0
pkgrel=1
pkgdesc="Desktop-style file explorer for the terminal"
arch=('x86_64' 'aarch64')
url="https://github.com/itsahmadawais/clidesk"
license=('MIT')
provides=('clidesk')

source_x86_64=("$url/releases/download/v$pkgver/clidesk-linux-amd64")
sha256sums_x86_64=('PASTE_SHA256_HERE')

source_aarch64=("$url/releases/download/v$pkgver/clidesk-linux-arm64")
sha256sums_aarch64=('PASTE_SHA256_HERE')

package() {
  install -Dm755 "clidesk-linux-*" "$pkgdir/usr/bin/clidesk"
}
```

---

## Release Checklist

When releasing a new version:

- [ ] Bump version in relevant files
- [ ] Push a git tag: `git tag v1.x.x && git push origin v1.x.x`
- [ ] GitHub Actions builds and uploads binaries automatically
- [ ] Update SHA256 hashes in:
  - [ ] Homebrew formula (`homebrew-tap` repo)
  - [ ] Scoop manifest (`scoop-bucket` repo)
  - [ ] Winget manifest (PR to `microsoft/winget-pkgs`)
  - [ ] AUR `PKGBUILD`

> With GoReleaser, the Homebrew and Scoop updates can be fully automated.

---

## Recommended Release Strategy

| Phase | Action |
|---|---|
| Now | `go install` — zero setup, targets Go developers |
| Soon | GitHub Releases + Actions — unlocks all non-Go users |
| Growth | Homebrew tap + Scoop bucket — one-line install on Mac/Windows |
| Mature | Winget submission + AUR — reaches the widest audience |
