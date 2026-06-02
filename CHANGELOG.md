# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Notes

- Ongoing development updates will be collected here until the next tagged release.
- This application currently is only being compiled for Windows, but the Go and Wails frameworks are multi-platform and this application could (in theory) be compiled on Linux and MacOS, though we haven't tested this yet.

## v0.2.1 - 2026-06-01

### Information

This is the first public release of this Markdown Reader application. It supports most CommonMark specs and GitHub Markdown syntax.

The application provides configuration options to enable/disable various Markdown syntax enhancements.

This release includes full configuration settings options and includes the ability for the user to select the proportional and monospace font of their choice and set the default font size for each. It has also had some refinements to the internal CSS to improve rendering, though more work needs to be done. Very basic printing support is available as well, though the CSS formatting for printing still needs tweaking.

This application currently is only being compiled for Windows, but the Go and Wails frameworks are multi-platform and this application could (in theory) be compiled on Linux and MacOS, though we haven't tested this yet.

> [!WARNING]
>
> This is still beta software and under heavy development work.

### Added

- Added a GitHub Actions workflow to test, build, and publish Windows releases from version tags.
- Added automated NSIS installer packaging, ZIP packaging, and release asset publishing.

### Changed

- Release assets are now published from CI using the tagged build output.
- ZIP archives and checksum files are now generated automatically during release builds.

### Fixed

- Improved CI reliability for Windows packaging by resolving frontend build ordering and NSIS installer setup issues.
