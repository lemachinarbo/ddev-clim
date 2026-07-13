# Changelog

## [0.5.1](https://github.com/lemachinarbo/ddev-clim/compare/v0.5.0...v0.5.1) (2026-07-13)


### Bug Fixes

* stop projects before starting during autostart ([518b888](https://github.com/lemachinarbo/ddev-clim/commit/518b888b69727d0b3e38c5588ff7237fd219cb20))

## [0.5.0](https://github.com/lemachinarbo/ddev-clim/compare/v0.4.2...v0.5.0) (2026-07-09)


### Features

* add automatic y/n confirmation prompt to power off DDEV on action failure ([9d74748](https://github.com/lemachinarbo/ddev-clim/commit/9d74748aa0795110b0d839fd85d333cced7dd98b))
* add ddev global status row and integrate poweroff stream logs ([6b4ef88](https://github.com/lemachinarbo/ddev-clim/commit/6b4ef8875752aee5b1f133978948df5a877a7b31))
* dynamically position details panel directly below the project list ([fe99f8d](https://github.com/lemachinarbo/ddev-clim/commit/fe99f8d69bb7b56321df8b09086f5c2d79bcc69c))
* handle unhealthy instances and add poweroff hotkey ([d3f2605](https://github.com/lemachinarbo/ddev-clim/commit/d3f26058c268f2f05b329a363e88780607e078fb))
* implement centralized project details and feedback panel with per-project error tracking ([8928579](https://github.com/lemachinarbo/ddev-clim/commit/892857930284e02245e9b0cbb7e9abb7d82adffd))
* implement Project Status Center with async container describe details and live stdout/stderr logs streaming ([fe832b2](https://github.com/lemachinarbo/ddev-clim/commit/fe832b20e068e8a992ff43af1d00f0e6b058bb9e))


### Bug Fixes

* display stopping... instead of starting... when toggling active instances ([c7e0512](https://github.com/lemachinarbo/ddev-clim/commit/c7e0512d4bd87f3635d5d9d1faa3a3dfe49d2797))
* format running feedback and truncate multi-line errors in details panel ([5ffae83](https://github.com/lemachinarbo/ddev-clim/commit/5ffae83c029cd721f6cc1a250f1978d9891b1d95))
* lock shortcuts to bottom of screen and override list status to failed when error is active ([2099356](https://github.com/lemachinarbo/ddev-clim/commit/20993568036aef1a3ba9e2d42844a493f468d4dc))
* only save project to running autostart config on successful action ([0051e06](https://github.com/lemachinarbo/ddev-clim/commit/0051e069b5ac063d4c75ee8a59eb2d52ac144beb))
* prevent TUI hang by closing stdout pipe when ddev exits, solving zombie child pipe inheritance ([eaa6cdd](https://github.com/lemachinarbo/ddev-clim/commit/eaa6cdda698e6b55b3dc85a628aa78fc1066b3ae))
* use os.Process.Wait instead of cmd.Wait to avoid circular deadlock over stdout pipe with child processes ([54c88c0](https://github.com/lemachinarbo/ddev-clim/commit/54c88c09d3aea67d20d693d555c6facca265a8f1))
* use raw reader to capture unbuffered progress logs directly, preventing newline line-buffering blocks ([d105a42](https://github.com/lemachinarbo/ddev-clim/commit/d105a42add82896da197b7b78783670eb07aef26))

## [0.4.2](https://github.com/lemachinarbo/ddev-clim/compare/v0.4.1...v0.4.2) (2026-05-27)


### Bug Fixes

* prevent memory wipe in TUI and add retry logic to autostart ([ecf7934](https://github.com/lemachinarbo/ddev-clim/commit/ecf7934422e2acda513acdf5259aca80c6d2f639))

## [0.4.1](https://github.com/lemachinarbo/ddev-clim/compare/v0.4.0...v0.4.1) (2026-05-24)


### Bug Fixes

* implement theme-aware ANSI colors for TUI ([47d0761](https://github.com/lemachinarbo/ddev-clim/commit/47d07610f5be20db970b8718687f2238991e340d))

## [0.4.0](https://github.com/lemachinarbo/ddev-clim/compare/v0.3.1...v0.4.0) (2026-05-24)


### Features

* refactor TUI to table layout and implement Zero Config mode ([0b610bd](https://github.com/lemachinarbo/ddev-clim/commit/0b610bde8b0239adf5dc2dc1c9d2a85fe8e12b29))

## [0.3.1](https://github.com/lemachinarbo/ddev-clim/compare/v0.3.0...v0.3.1) (2026-05-23)


### Bug Fixes

* update systemd path to /usr/bin for AUR and update README ([fe0b979](https://github.com/lemachinarbo/ddev-clim/commit/fe0b9795ac71579e20ef1eaba3ab78a1b26f8d2f))
* update systemd service to use /usr/bin path for AUR compatibility ([f3ae675](https://github.com/lemachinarbo/ddev-clim/commit/f3ae6751536c57795d6b75436fe3b41ffca1ba18))

## [0.3.0](https://github.com/lemachinarbo/ddev-clim/compare/v0.2.1...v0.3.0) (2026-05-23)


### Features

* initial release of DDEV CLInstance Manager (ddev-clim) ([0b475a0](https://github.com/lemachinarbo/ddev-clim/commit/0b475a0678e77ca3d0bffdc4a7fbf14d1dc4f5b6))


### Bug Fixes

* correct source directory path in PKGBUILD ([2f9b1d5](https://github.com/lemachinarbo/ddev-clim/commit/2f9b1d57b096b85a878704b1253e7b678193b578))
* remove package-name from release-please to ensure valid semver tags ([31f87bd](https://github.com/lemachinarbo/ddev-clim/commit/31f87bd7b01ffa85af960f5bd4a02eee4809cb23))
* update release workflow to trigger on master branch ([44922de](https://github.com/lemachinarbo/ddev-clim/commit/44922de69854de6205e4f10b6f5c8c0d9012aa18))

## [0.2.1](https://github.com/lemachinarbo/ddev-clim/compare/ddev-clim-v0.2.0...ddev-clim-v0.2.1) (2026-05-23)


### Bug Fixes

* correct source directory path in PKGBUILD ([2f9b1d5](https://github.com/lemachinarbo/ddev-clim/commit/2f9b1d57b096b85a878704b1253e7b678193b578))

## [0.2.0](https://github.com/lemachinarbo/ddev-clim/compare/ddev-clim-v0.1.0...ddev-clim-v0.2.0) (2026-05-23)


### Features

* initial release of DDEV CLInstance Manager (ddev-clim) ([0b475a0](https://github.com/lemachinarbo/ddev-clim/commit/0b475a0678e77ca3d0bffdc4a7fbf14d1dc4f5b6))


### Bug Fixes

* update release workflow to trigger on master branch ([44922de](https://github.com/lemachinarbo/ddev-clim/commit/44922de69854de6205e4f10b6f5c8c0d9012aa18))
