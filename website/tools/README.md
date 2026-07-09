# Downloadable tools

Copies of the local helper scripts, served from the site so the Packager admin
tab can offer them as downloads (the deploy only bundles `website/`, so the
real scripts under repo-root `scripts/` are not reachable in production).

- `package-modpack.ps1` — builds a `.polypack` from a profile folder.
- `slime-lib.ps1` — obfuscation helper that `package-modpack.ps1` dot-sources.
  Keep both in the **same folder** when you run the packager.

**These are generated copies — do not edit them here.** Edit the originals in
`scripts/` instead; `scripts/package-website.ps1` refreshes these on every
package so the deployed copies never go stale.
