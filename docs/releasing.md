# Releasing

Lightweight process for tagging **semver** releases. The integration line is **`develop`**; **`main`** tracks what end users should consider stable (see [CONTRIBUTING.md](../CONTRIBUTING.md)).

## Preconditions

1. **`develop` is green:** required GitHub Actions checks pass.
2. **Local parity (recommended):** from the repo root, `make ci` (or `./scripts/ci.sh`) succeeds.
3. **Changelog:** move items from `[Unreleased]` under a new `## [x.y.z] — YYYY-MM-DD` section in [CHANGELOG.md](../CHANGELOG.md), then leave empty `[Unreleased]` scaffolding.

## Steps

1. Merge or fast-forward **`main`** from **`develop`** (per project policy).
2. **Tag** the release commit on `main`:

   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

   Use **annotated** tags so `git describe` and tooling see a real release.

3. **Module consumers** can install with:

   ```bash
   go install github.com/ipolyzos/service-manager-read-mcp/cmd/service-manager-read-mcp@v0.1.0
   ```

4. **Optional:** create a [GitHub Release](https://github.com/ipolyzos/service-manager-read-mcp/releases) from the tag and paste the changelog section for that version.

## After the first tag

Update the footer links at the bottom of `CHANGELOG.md` if compare URLs need adjusting. For security-sensitive fixes, follow [SECURITY.md](../SECURITY.md).
