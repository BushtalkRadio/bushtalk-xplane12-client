# Bushtalk Radio Companion

Standalone desktop app for X-Plane 12 flight tracking. Runs alongside X-Plane as a separate window.

## Requirements

- X-Plane 12.1.1 or later (includes the Web API)
- A Bushtalk Radio account

## Installation

1. Download the appropriate binary for your platform from the [Releases](https://github.com/BushtalkRadio/bushtalk-xplane12-client/releases) page
2. Place it anywhere on your computer
3. Run the application

No additional setup required - X-Plane's Web API runs automatically on port 8086.

**Note:** Your antivirus may flag this program because the .exe is not signed with a security certificate. This is a false positive. The source code is publicly available if you prefer to compile it yourself.

## Usage

1. Start X-Plane 12
2. Run the Bushtalk Radio Companion
3. Enter your Bushtalk Radio username and password
4. Click Login

Once connected, your position is sent to Bushtalk Radio every 5 seconds and appears on the [live map](https://bushtalkradio.com/map).

## Configuration

Settings are stored in `config.json`:

| Platform | Location |
|----------|----------|
| Windows | `%APPDATA%\BushtalkRadio\config.json` |
| macOS | `~/Library/Application Support/BushtalkRadio/config.json` |
| Linux | `~/.config/bushtalkradio/config.json` |

## Troubleshooting

### "X-Plane: Disconnected"

- Ensure X-Plane 12 is running
- Check that X-Plane's Web API is enabled (it is by default)
- Try restarting X-Plane

### Port Conflicts

Port 8086 is occasionally used by other software. If you have conflicts:

1. In X-Plane: Settings > Data Output > Web Interface
2. Change the port number
3. Update the port in the companion's settings

## Building from Source

Requires Go 1.21+ and Fyne dependencies.

```bash
cd companion

# Install dependencies
go mod download

# Build for current platform
go build -o bushtalk-companion .

# Cross-compile (Windows from Linux/WSL)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
    go build -ldflags="-H windowsgui" -o bushtalk-companion.exe .
```

### Fyne Dependencies

See [Fyne Getting Started](https://developer.fyne.io/started/) for platform-specific requirements.

## Support

- Website: https://bushtalkradio.com
- Discord: https://discord.com/invite/ZcGgw9mUqA
- Email: admin@bushtalkradio.com
