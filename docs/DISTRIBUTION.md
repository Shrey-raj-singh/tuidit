# Distribution guide

How to publish tuidit to GitHub Releases, Winget, Homebrew, and (optionally) APT.

## 1. GitHub Releases (primary)

1. Bump version in code/docs if needed (e.g. `Version` in `cmd/editor/main.go` for `tuidit --version`).
2. Commit, then create and push a tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. The [release workflow](.github/workflows/release.yml) builds Linux (amd64/arm64), Windows (amd64), and macOS (amd64/arm64) and attaches them to the GitHub Release.

## 2. Winget (Windows)

Manifests live in this repo under `winget/`. To add a **new version** to the [Windows Package Manager](https://github.com/microsoft/winget-pkgs) repo:

1. After the GitHub Release is published, download `tuidit-windows-amd64.exe` and compute SHA256:
   ```powershell
   (Get-FileHash -Algorithm SHA256 .\tuidit-windows-amd64.exe).Hash
   ```
2. Copy the `winget/` folder and create a versioned directory in winget-pkgs:
   - In [winget-pkgs](https://github.com/microsoft/winget-pkgs): `manifests/s/ShreyRajSingh/Tuidit/<PackageVersion>/`
   - Put the three YAML files there: **version**, **locale.en-US**, and **installer**. In the **installer** file set:
     - `PackageVersion` to the new version (e.g. `1.0.0`)
     - `InstallerUrl` to `https://github.com/Shrey-raj-singh/tuidit/releases/download/v<version>/tuidit-windows-amd64.exe`
     - `InstallerSha256` to the value from step 1.
   - In the **version** and **locale.en-US** files set `PackageVersion` to the same value.
   - File names must match winget rules: `ShreyRajSingh.Tuidit.version.yaml`, `ShreyRajSingh.Tuidit.locale.en-US.yaml`, `ShreyRajSingh.Tuidit.installer.yaml` (not `defaultLocale.yaml`).
3. Validate and submit:
   ```powershell
   winget validate --manifest manifests\s\ShreyRajSingh\Tuidit\<version>\
   ```
   Then open a PR to [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs) with the new manifest folder.

Alternatively use [wingetcreate](https://github.com/microsoft/winget-create) to generate/update manifests from a release URL.

## 3. Homebrew (macOS / Linux)

The formula is in `Formula/tuidit.rb`. To publish via a **tap**:

1. Create a repo named `homebrew-tuidit` (or keep the formula in the main repo under `Formula/`).
2. For each new release, update in `Formula/tuidit.rb`:
   - `version`
   - All `url` lines to the new tag (e.g. `v1.0.0`).
   - All `sha256` values. After the release is published:
     ```bash
     curl -sL -o /tmp/tuidit-darwin-arm64 "https://github.com/Shrey-raj-singh/tuidit/releases/download/v1.0.0/tuidit-darwin-arm64"
     shasum -a 256 /tmp/tuidit-darwin-arm64
     ```
     Repeat for each asset (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64).
3. If using a separate tap repo, push the updated formula. Users install with:
   ```bash
   brew tap Shrey-raj-singh/tuidit https://github.com/Shrey-raj-singh/tuidit
   brew install tuidit
   ```
   If the formula lives in the main repo under `Formula/`:
   ```bash
   brew install Shrey-raj-singh/tuidit/tuidit
   ```

## 4. APT (Debian/Ubuntu)

The release workflow already:

1. Builds **.deb** packages (amd64 and arm64) with **nfpm** (`nfpm.yaml`) and uploads them to the GitHub Release.
2. Generates an **APT repo** (pool + dists) and deploys it to **GitHub Pages** via the “Upload APT repo for GitHub Pages” step and the `deploy-pages` job.

**Requirements:**

- In the repo: **Settings → Pages → Build and deployment**: set **Source** to **GitHub Actions**.
- After the first tagged release, the APT repo will be at `https://<owner>.github.io/tuidit/` (e.g. `https://shrey-raj-singh.github.io/tuidit/`).

Users add the repo and install with:

```bash
echo "deb [trusted=yes] https://shrey-raj-singh.github.io/tuidit/ stable main" | sudo tee /etc/apt/sources.list.d/tuidit.list
sudo apt update
sudo apt install tuidit
```

(We use `[trusted=yes]` because the repo is not GPG-signed. To sign the Release file, add GPG signing in the “Generate APT repo” step and document the key install.)
