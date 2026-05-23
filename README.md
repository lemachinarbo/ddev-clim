# DDEV CLI Manager

An advanced, beautiful Terminal UI for managing your DDEV instances, built with Go and Bubble Tea.

## Features

- **TUI Interface**: List all DDEV projects with status indicators.
- **Toggle On/Off**: Quickly start or stop projects with `Enter` or `Space`.
- **Session Memory**: Automatically remembers which projects were running.
- **Autostart**: Restore your last session with a single command.

## Installation

```bash
# Build the binary
go build -o ddev-clim main.go

# (Optional) Move to your bin folder
mv ddev-clim ~/.local/bin/
```

## Usage

### Interactive TUI
Just run the command to open the manager:
```bash
ddev-clim
```
- `j/k` or `arrows`: Navigate
- `Enter` or `Space`: Toggle project (Start/Stop)
- `/`: Search/Filter projects
- `q` or `Ctrl+C`: Quit

### Scan specific folder
To scan a specific folder (like `~/projects`) and persist this setting:
```bash
ddev-clim --path ~/projects
```
Once set, the TUI will always scan this folder for DDEV instances.

### Autostart
To restore the projects that were running when you last used the TUI:
```bash
ddev-clim autostart
```

## System Integration (Auto-start on Boot)

Since you are using Hyprland/Omarchy, you can add this to your `hyprland.conf`:
```conf
exec-once = ddev-clim autostart
```
Or add it to your shell's profile/rc if you want it to run when you open a terminal.

## Persistence
The session state is stored in `~/.config/ddev-clim/config.json`.
