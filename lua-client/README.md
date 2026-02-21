# Bushtalk Radio Plugin

In-sim flight tracker for X-Plane 12. Runs inside X-Plane via FlyWithLua.

## Requirements

- X-Plane 12
- [FlyWithLua NG](https://forums.x-plane.org/index.php?/files/file/82888-flywithlua-ng-next-generation-plus-edition-for-x-plane-12/)

## Installation

1. **Install FlyWithLua NG** (if not already installed):
   - Download from [X-Plane.org](https://forums.x-plane.org/index.php?/files/file/82888-flywithlua-ng-next-generation-plus-edition-for-x-plane-12/)
   - Extract to `X-Plane 12/Resources/plugins/`

2. **Install Bushtalk Radio Plugin**:
   - Download `bushtalk.lua`
   - Copy to `X-Plane 12/Resources/plugins/FlyWithLua/Scripts/`

3. **Approve the script** (first run only):
   - FlyWithLua may quarantine new scripts
   - Go to `Plugins > FlyWithLua > Quarantine` and approve `bushtalk.lua`

4. **Start X-Plane** - the plugin loads automatically

## Usage

1. A window appears when X-Plane starts
2. Log in with your Bushtalk Radio account
3. Tracking starts automatically once logged in
4. Your flight appears on the [live map](https://bushtalkradio.com/map)

After your first login, credentials are saved and you'll be logged in automatically on future flights.

## Troubleshooting

### "LuaSocket not available"

FlyWithLua NG should include LuaSocket. If you see this error:

1. Make sure you have FlyWithLua **NG** (Next Generation), not the older version
2. Check that these files exist in `FlyWithLua/Modules/`:
   - `socket.lua`
   - `socket/` folder with `http.lua`, `core.dll` (or `.so`/`.dylib`)

### Window doesn't appear

- Check `Plugins > FlyWithLua > FlyWithLua Macros > Bushtalk Radio`
- Check X-Plane's `Log.txt` for errors

## Config Location

```
X-Plane 12/Output/preferences/bushtalk_config.txt
```

## Support

- Website: https://bushtalkradio.com
- Discord: https://discord.com/invite/ZcGgw9mUqA
- Email: admin@bushtalkradio.com
