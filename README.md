# DDEV CLInstance Manager

I always forgot to start the ddev instance of the projects I'm working on, so "I" 🤖 created this tool.
What is for? You can toggle on and off all the ddev instances of the projects you are working on and they will start on boot (if you want).

## Installation

### Arch Linux (AUR) - Recommended
Install using your favorite AUR helper:
```bash
yay -S ddev-clim
# or
paru -S ddev-clim
```

### From Source
```bash
# Clone the repository
git clone https://github.com/lemachinarbo/ddev-clim.git
cd ddev-clim

# Build and install
go install .
```
Ensure your `$GOPATH/bin` is in your `$PATH`.

## Usage

### Interactive TUI
Just run the command to open the manager:
```bash
ddev-clim
```
- `j/k` or `arrows`: Navigate
- `Enter` or `Space`: Toggle project (Start/Stop)
- `p`: Open folder picker to change scan directory
  - `.`: Toggle hidden files/folders in the picker
- `/`: Search/Filter projects
- `q` or `Ctrl+C`: Quit

### Scan specific folder
You can change the scan path interactively within the TUI by pressing `p`. Alternatively, set it via CLI:
```bash
ddev-clim --path ~/projects
```

### Autostart
To restore the projects that were running when you last used the TUI:
```bash
ddev-clim autostart
```

## System Integration (Auto-start on Boot)

### Using Systemd (Recommended for Linux)
The AUR package automatically installs the service file. You just need to enable it:
```bash
systemctl --user enable --now ddev-clim.service
```

### Using Hyprland (Omarchy)
Add this to your `hyprland.conf`:
```conf
exec-once = ddev-clim autostart
```

## Development & Releases
This project uses **Conventional Commits** and **GoReleaser** paired with Google's **release-please**.
- To trigger a new release, use commit prefixes like `feat:` or `fix:`.
- Detailed instructions for maintainers can be found in [AGENTS.md](./AGENTS.md).

## Persistence
The session state is stored in `~/.config/ddev-clim/config.json`.
