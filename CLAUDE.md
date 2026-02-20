# Bushtalk Radio X-Plane Client

Go desktop app that connects X-Plane 12 to Bushtalk Radio for live flight tracking.

## Stack

- **Go 1.21+** - Single binary, cross-platform
- **Fyne v2.4** - Cross-platform GUI toolkit
- **gorilla/websocket** - WebSocket client for X-Plane API

## Architecture

```
main.go              # Entry point, goroutine orchestration
theme.go             # Custom Fyne theme (dark, matches website)
resources.go         # Embedded icon (go:embed)
console_windows.go   # Windows console hide/show (syscall)
console_other.go     # No-op for non-Windows

config/
  config.go          # Load/save config from %APPDATA%\BushtalkRadio\

bushtalk/
  client.go          # HTTP client for /api/authenticate + /api/track

xplane/
  client.go          # WebSocket client, position updates
  datarefs.go        # REST dataref ID resolution, tail number decode

ui/
  login.go           # Login form window
  status.go          # Status window with flight data
```

## X-Plane Web API (port 8086)

X-Plane 12.1.1+ has a built-in Web API. Two-step connection:

1. **REST** - Resolve dataref names to session-specific IDs:
   ```
   GET http://localhost:8086/api/v3/datarefs?filter[name]=sim/flightmodel/position/latitude
   → {"data":[{"id":123456789,"name":"...","value_type":"double"}]}
   ```

2. **WebSocket** - Subscribe using numeric IDs:
   ```json
   {"req_id":1,"type":"dataref_subscribe_values","params":{"datarefs":[{"id":123456789}]}}
   ```

3. **Streaming** - X-Plane pushes at ~10Hz:
   ```json
   {"data":{"123456789":47.804604},"type":"dataref_update_values"}
   ```

**Important**: IDs change every X-Plane session - must re-resolve on reconnect.

## Datarefs Used

| Dataref | Type | Description |
|---------|------|-------------|
| sim/flightmodel/position/latitude | double | Latitude |
| sim/flightmodel/position/longitude | double | Longitude |
| sim/flightmodel/position/y_agl | float | Altitude AGL (meters) |
| sim/flightmodel/position/groundspeed | float | Ground speed (m/s) |
| sim/flightmodel/position/mag_psi | float | Magnetic heading |
| sim/aircraft/view/acf_tailnum | data | Tail number (byte array) |

## Bushtalk API

Headers sent on all requests:
- `X-BTR-CLIENT-NAME: XPLANE-12`
- `X-BTR-CLIENT-VERSION: 1.0.0`

**POST /api/authenticate** - Login, returns `id_token`

**POST /api/track** - Send position (every 5 seconds):
```json
{
  "PLANE_LATITUDE": -33.86,
  "PLANE_LONGITUDE": 151.21,
  "ALTITUDE_ABOVE_GROUND": 500,
  "GROUND_VELOCITY": 120,
  "MAGNETIC_COMPASS": 270,
  "ATC_ID": "VH-ABC",
  "SIM_ON_GROUND": false
}
```

## Config File

Location: `%APPDATA%\BushtalkRadio\config.json` (Windows)

```json
{
  "username": "pilot123",
  "api_token": "eyJ...",
  "api_url": "https://bushtalkradio.com",
  "xplane_port": 8086,
  "show_console": false
}
```

## Commands

### Setup (WSL2)

```bash
# Install Go and MinGW cross-compiler
sudo apt install -y golang-go gcc-mingw-w64-x86-64

# Install dependencies
cd clients/xplane
go mod tidy
```

### Building

```bash
# Development build (console visible for debugging)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
    go build -o bushtalk-xplane.exe .

# Release build (console hidden)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
    go build -ldflags="-H windowsgui" -o bushtalk-xplane.exe .

# Copy to Windows desktop (WSL2)
cp bushtalk-xplane.exe "/mnt/c/Users/patri/OneDrive/Desktop/"
```

### Testing X-Plane API (with X-Plane running)

```bash
# Check if API is running
curl -s "http://localhost:8086/"

# List all datarefs (large response)
curl -s "http://localhost:8086/api/v3/datarefs" | head -200

# Resolve specific dataref to session ID
curl -s "http://localhost:8086/api/v3/datarefs?filter\[name\]=sim/flightmodel/position/latitude"
curl -s "http://localhost:8086/api/v3/datarefs?filter\[name\]=sim/flightmodel/position/longitude"
curl -s "http://localhost:8086/api/v3/datarefs?filter\[name\]=sim/flightmodel/position/y_agl"
curl -s "http://localhost:8086/api/v3/datarefs?filter\[name\]=sim/flightmodel/position/groundspeed"
curl -s "http://localhost:8086/api/v3/datarefs?filter\[name\]=sim/flightmodel/position/mag_psi"
curl -s "http://localhost:8086/api/v3/datarefs?filter\[name\]=sim/aircraft/view/acf_tailnum"
```

### Running from PowerShell (to see debug logs)

```powershell
cd ~\OneDrive\Desktop
.\bushtalk-xplane.exe
```

## Unit Conversions

- Altitude: meters → feet (* 3.28084)
- Speed: m/s → knots (* 1.94384)
- On ground: AGL < 1 meter

## Known Issues

- Tail number dataref (`acf_tailnum`) may not stream values for all aircraft
- Console visibility toggle requires app restart
