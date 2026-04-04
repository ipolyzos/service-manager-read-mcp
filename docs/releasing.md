# Releasing

Lightweight process for tagging **semver** releases. The integration line is **`develop`**; **`main`** tracks what end users should consider stable (see [contributing.md](contributing.md)).

## Preconditions

1. **`develop` is green:** required GitHub Actions checks pass.
2. **Local parity (recommended):** from the repo root, `make ci` (or `./scripts/ci.sh`) succeeds.
3. **Changelog:** move items from `[Unreleased]` under a new `## [x.y.z] — YYYY-MM-DD` section in [CHANGELOG.md](CHANGELOG.md), then leave empty `[Unreleased]` scaffolding.

## Steps

1. Promote **`develop` → `main`** via **GitHub Pull Request** (base **`main`**, compare **`develop`**). Merge only after required checks are green. **Do not** `git push` straight to **`main`** to skip CI.
2. Locally: `git checkout main && git pull` so you tag the same commit GitHub merged.
3. **Tag** the release commit on `main`:

   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

   Use **annotated** tags so `git describe` and tooling see a real release.

4. **Module consumers** can install with:

   ```bash
   go install github.com/ipolyzos/service-manager-read-mcp/cmd/service-manager-read-mcp@vX.Y.Z
   ```

5. **Optional:** create a [GitHub Release](https://github.com/ipolyzos/service-manager-read-mcp/releases) from the tag and paste the changelog section for that version.

## After the first tag

Update the footer links at the bottom of `CHANGELOG.md` if compare URLs need adjusting. For security-sensitive fixes, follow [security.md](security.md).
