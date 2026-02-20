# Bushtalk Radio X-Plane Client

A lightweight flight tracker that connects X-Plane 12 to [Bushtalk Radio](https://bushtalkradio.com), allowing you to share your flights with the aviation community.

## Requirements

- **X-Plane 12.1.1 or later** (includes the Web API)
- A Bushtalk Radio account

## Installation

1. Download the appropriate binary for your platform from the releases page
2. Place it anywhere on your computer
3. Run the application

No additional setup is required - X-Plane's Web API runs automatically on port 8086.

## Usage

1. Start X-Plane 12
2. Run the Bushtalk Radio client
3. Enter your Bushtalk Radio username and password
4. Check "Remember me" to save your credentials
5. Click Login

Once connected, your position will be sent to Bushtalk Radio every 5 seconds and appear on the live map.

## Configuration

The client stores settings in `config.json` in the same directory as the executable.

### Advanced Settings

Click "Advanced Settings" in the login window to configure:

- **X-Plane Port**: Default is 8086. Change if you've configured X-Plane to use a different port, or if another application is using 8086.
- **API URL**: For development/testing purposes only.

## Troubleshooting

### "X-Plane: Disconnected"

- Ensure X-Plane 12 is running
- Check that X-Plane's Web API is enabled (it is by default)
- Try restarting X-Plane

### Port Conflicts

Port 8086 is occasionally used by other software (InfluxDB, RGB controllers, etc.). If you have conflicts:

1. In X-Plane, go to Settings → Data Output → Web Interface
2. Change the port number
3. Update the X-Plane Port in the client's Advanced Settings

### Firewall Issues

If you're having connection issues:

- Ensure your firewall allows localhost connections
- The client only connects to `localhost` - no incoming connections are required

## Building from Source

Requires Go 1.21+ and Fyne dependencies.

```bash
# Install dependencies
go mod download

# Build for current platform
go build -o bushtalk-xplane .

# Cross-compile all platforms
./build/build.sh
```

### Fyne Dependencies

Fyne requires native graphics libraries. See [Fyne Getting Started](https://developer.fyne.io/started/) for platform-specific requirements.

**Linux:**
```bash
sudo apt-get install gcc libgl1-mesa-dev xorg-dev
```

**macOS:**
```bash
xcode-select --install
```

**Windows:**
- Install MinGW-w64 or use MSYS2

## How It Works

1. The client connects to X-Plane's local Web API (HTTP + WebSocket on port 8086)
2. It subscribes to flight data: position, altitude, heading, groundspeed, tail number
3. Every 5 seconds, it sends your current position to Bushtalk Radio
4. Your flight appears on the live map at bushtalkradio.com

## Privacy

- The client only sends: latitude, longitude, altitude, heading, speed, and aircraft tail number
- No other X-Plane data is transmitted
- Your Bushtalk Radio credentials are stored locally if you select "Remember me"

## License

MIT License - See LICENSE file for details.
