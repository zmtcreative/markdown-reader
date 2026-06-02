<!-- markdownlint-disable MD024 -->

# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Notes

- Ongoing development updates will be collected here until the next tagged release.

## v0.2.3 - 2026-06-02

This is the first public release of this Markdown Reader application. It supports most CommonMark specs and GitHub Markdown syntax.

The application provides configuration options to enable/disable various Markdown syntax enhancements.

This release includes full configuration settings options and includes the ability for the user to select the proportional and monospace font of their choice and set the default font size for each. It has also had some refinements to the internal CSS to improve rendering, though more work needs to be done. Very basic printing support is available as well, though the CSS formatting for printing still needs tweaking.

The application has been compiled and tested for Windows, and the tests and build completed successfully on both WSL2 (Ubuntu-26.04) and within the GitHub CI workflow (to generate a `.tar.gz` package containing the Linux executable). The Go and Wails frameworks are multi-platfor and this application could (in theory) be compiled on MacOS as well, but this hasn't been tested at this time.

### Changed

- Release assets for both Windows and Linux are being provided for installing on each platform
- The NSIS installer is compressed into a `.zip` file to make downloads on Windows slightly less annoying
- The Linux executable is packaged into a `.tar.gz` file and can be extracted and run from any location you choose

## v0.2.2 - 2026-06-01

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
