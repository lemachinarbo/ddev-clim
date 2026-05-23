# Agent & Maintainer Instructions

This project uses a dual-remote setup to manage both the source code (GitHub) and the Arch Linux package metadata (AUR).

## Remote Configuration
- **origin**: `https://github.com/lemachinarbo/ddev-clim.git` (Source code & Releases)
- **aur**: `aur@aur.archlinux.org:ddev-clim.git` (AUR Package Metadata)

## Release Workflow

### 1. Source Release (GitHub)
We use **Conventional Commits** and **release-please**.
1. Commit changes using prefixes: `feat:`, `fix:`, `chore:`, etc.
2. Push to GitHub: `git push origin main`.
3. Go to GitHub and merge the "Release PR" created by `release-please`.
4. This will automatically create a tag (e.g., `v0.2.0`) and publish binaries via GoReleaser.

### 2. Package Release (AUR)
Once the GitHub tag is live:
1. Update `pkgver` in `PKGBUILD` to match the new tag.
2. Update checksums: `makepkg -g >> PKGBUILD` (and remove the old `sha256sums` line).
3. Update metadata: `makepkg --printsrcinfo > .SRCINFO`.
4. Commit these two files: `git add PKGBUILD .SRCINFO && git commit -m "Release vX.Y.Z"`.
5. **Push to AUR**:
   ```bash
   git push aur main:master
   ```
   *(Note: AUR only accepts the `master` branch, so we push our `main` to their `master`)*

## Development Notes
- TUI uses `charmbracelet/bubbletea`.
- Session state is stored in `~/.config/ddev-clim/config.json`.
- The interactive folder picker is accessed with `p` and toggles hidden files with `.`.
