# DDEV CLI Manager Plan

## Objective
Create an advanced, beautiful CLI interface in Go using Bubble Tea to manage DDEV instances. The tool will list projects, allow starting/stopping them interactively, and implement a session-based memory to automatically start previously running projects upon execution.

## Key Files & Context
- `main.go`: Entry point and Cobra CLI setup.
- `ui/`: Bubble Tea TUI components (model, update, view).
- `ddev/`: Wrapper around `ddev` shell commands (parsing `ddev list -j`, executing `start`/`stop`).
- `config/`: JSON configuration management for session memory.

## Implementation Steps
1. **CLI Structure setup**:
   - Initialize Go modules (`go mod init github.com/username/ddev-clim`).
   - Setup Cobra CLI framework with root command (TUI) and `auto-start` subcommand.
2. **DDEV Integration**:
   - Create a Go package to execute `ddev list -j` and parse the JSON output to get projects (`name`, `status`, `approot`).
   - Implement functions for `ddev start` and `ddev stop` using `os/exec`, running them in the respective `approot` directories.
3. **Session Memory (Persistence)**:
   - Create a configuration manager that reads/writes to `~/.config/ddev-clim/config.json`.
   - Store an array of currently "running" or "desired" project names.
   - Update this file whenever a project is toggled via the UI.
4. **Beautiful TUI (Bubble Tea)**:
   - Implement a list view (using `bubbles/list` or `bubbles/table`) to display projects and their status.
   - Map keybindings: `Enter` or `Space` to toggle status (start/stop), `q` to quit.
   - Add status indicators (e.g., green for running, gray for stopped).
5. **Auto-Start Command**:
   - Implement the `auto-start` subcommand that reads the config file and executes `ddev start` for all remembered projects.

## Verification & Testing
- Run the TUI and ensure it correctly lists all DDEV projects.
- Toggle a project and verify `ddev start`/`stop` is executed and UI updates.
- Verify that toggling a project updates the `config.json`.
- Run `ddev-clim auto-start` and verify it restores the last session's state.