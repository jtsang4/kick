# Kick

Kick is a command-line tool for personal use.

## Features

- **SSH Configuration**: One-click SSH security hardening
  - Enable key-based authentication
  - Disable password login
  - Configure security parameters
  - Automatically add public keys

## Installation

### From Binary

Download the appropriate binary for your system from the [Releases](https://github.com/jtsang4/kick/releases) page.

### Build from Source

```bash
git clone https://github.com/jtsang4/kick.git
cd kick
go build
```

## Usage

### SSH Configuration

Requires root privileges:

```bash
sudo kick ssh
```

After running the command, follow the interactive prompts to enter your SSH public key.

## System Requirements

- Supports Linux, macOS, and Windows
- Root privileges required for SSH configuration

## Tech Stack

- Go language
- [Cobra](https://github.com/spf13/cobra) for building CLI interfaces
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for interactive TUI

## License

[MIT](LICENSE)
