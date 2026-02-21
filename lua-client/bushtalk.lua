-- Bushtalk Radio for X-Plane 12
-- FlyWithLua script for live flight tracking
-- Drop this file into: X-Plane 12/Resources/plugins/FlyWithLua/Scripts/

local VERSION = "1.0.0"
local API_URL = "https://bushtalkradio.com"
local TRACK_INTERVAL = 5 -- seconds

-- Try to load LuaSocket for HTTP requests
local http_available = false
local http, ltn12, json

local function load_dependencies()
    local ok, result = pcall(require, "socket.http")
    if ok then
        http = result
        http.TIMEOUT = 10
        ltn12 = require("ltn12")
        http_available = true
    else
        logMsg("Bushtalk: LuaSocket not available - " .. tostring(result))
    end
end

load_dependencies()

-- Simple JSON encoder (FlyWithLua doesn't include json by default)
local function json_encode(tbl)
    local function encode_value(val)
        local t = type(val)
        if t == "string" then
            return '"' .. val:gsub('\\', '\\\\'):gsub('"', '\\"'):gsub('\n', '\\n') .. '"'
        elseif t == "number" then
            return tostring(val)
        elseif t == "boolean" then
            return val and "true" or "false"
        elseif t == "table" then
            local parts = {}
            local is_array = #val > 0
            if is_array then
                for _, v in ipairs(val) do
                    table.insert(parts, encode_value(v))
                end
                return "[" .. table.concat(parts, ",") .. "]"
            else
                for k, v in pairs(val) do
                    table.insert(parts, '"' .. tostring(k) .. '":' .. encode_value(v))
                end
                return "{" .. table.concat(parts, ",") .. "}"
            end
        end
        return "null"
    end
    return encode_value(tbl)
end

-- Simple JSON decoder for auth response
local function json_decode(str)
    -- Very basic parser for simple objects
    local result = {}
    for key, value in str:gmatch('"([^"]+)"%s*:%s*"?([^",}]+)"?') do
        if value == "true" then
            result[key] = true
        elseif value == "false" then
            result[key] = false
        elseif tonumber(value) then
            result[key] = tonumber(value)
        else
            result[key] = value
        end
    end
    return result
end

-- Config file handling
local config_path = SYSTEM_DIRECTORY .. "Output/preferences/bushtalk_config.txt"

local config = {
    username = "",
    token = "",
}

local function save_config()
    local f = io.open(config_path, "w")
    if f then
        f:write("username=" .. config.username .. "\n")
        f:write("token=" .. config.token .. "\n")
        f:close()
    end
end

local function load_config()
    local f = io.open(config_path, "r")
    if f then
        for line in f:lines() do
            local key, value = line:match("^(%w+)=(.*)$")
            if key and value then
                config[key] = value
            end
        end
        f:close()
    end
end

load_config()

-- Datarefs
local dr_latitude = dataref_table("sim/flightmodel/position/latitude")
local dr_longitude = dataref_table("sim/flightmodel/position/longitude")
local dr_altitude_agl = dataref_table("sim/flightmodel/position/y_agl")
local dr_groundspeed = dataref_table("sim/flightmodel/position/groundspeed")
local dr_heading = dataref_table("sim/flightmodel/position/mag_psi")
local dr_tailnum = dataref_table("sim/aircraft/view/acf_tailnum")

-- State
local state = {
    logged_in = config.token ~= "",
    last_send = 0,
    last_error = "",
    last_position = nil,
    show_window = true,
}

-- Input buffers for imgui
local input_username = config.username
local input_password = ""

-- Get tail number from byte array dataref
local function get_tail_number()
    local tail = ""
    for i = 0, 39 do
        local byte = dr_tailnum[i]
        if byte == nil then break end
        -- Handle both number and string returns
        if type(byte) == "number" then
            if byte == 0 then break end
            tail = tail .. string.char(byte)
        elseif type(byte) == "string" then
            if byte == "" or byte == "\0" then break end
            tail = tail .. byte
        end
    end
    if tail == "" then tail = "UNKNOWN" end
    return tail
end

-- HTTP POST request
local function http_post(url, body, headers)
    if not http_available then
        return nil, "LuaSocket not available"
    end

    local response_body = {}
    local result, status_code, response_headers = http.request{
        url = url,
        method = "POST",
        headers = headers,
        source = ltn12.source.string(body),
        sink = ltn12.sink.table(response_body),
    }

    if not result then
        return nil, status_code or "Request failed"
    end

    return table.concat(response_body), status_code
end

-- Authenticate with Bushtalk API
local function authenticate(username, password)
    local body = json_encode({
        username = username,
        password = password,
    })

    local headers = {
        ["Content-Type"] = "application/json",
        ["Content-Length"] = tostring(#body),
        ["X-BTR-CLIENT-NAME"] = "XPLANE-12-LUA",
        ["X-BTR-CLIENT-VERSION"] = VERSION,
    }

    local response, status = http_post(API_URL .. "/api/authenticate", body, headers)

    if not response then
        return false, status
    end

    if status ~= 200 then
        return false, "Authentication failed (status " .. status .. ")"
    end

    local data = json_decode(response)
    if data.id_token then
        config.username = username
        config.token = data.id_token
        save_config()
        state.logged_in = true
        state.last_error = ""
        return true, nil
    end

    return false, "Invalid response"
end

-- Send position to Bushtalk API
local function send_position()
    if not state.logged_in or config.token == "" then
        return false, "Not logged in"
    end

    local lat = dr_latitude[0]
    local lon = dr_longitude[0]

    -- Skip if no valid position
    if lat == 0 and lon == 0 then
        return false, "No position data"
    end

    local alt_meters = dr_altitude_agl[0]
    local speed_ms = dr_groundspeed[0]

    local payload = {
        PLANE_LATITUDE = lat,
        PLANE_LONGITUDE = lon,
        ALTITUDE_ABOVE_GROUND = alt_meters * 3.28084,  -- meters to feet
        GROUND_VELOCITY = speed_ms * 1.94384,          -- m/s to knots
        MAGNETIC_COMPASS = dr_heading[0],
        ATC_ID = get_tail_number(),
        SIM_ON_GROUND = alt_meters < 1.0,
    }

    state.last_position = payload

    local body = json_encode(payload)

    local headers = {
        ["Content-Type"] = "application/json",
        ["Content-Length"] = tostring(#body),
        ["Authorization"] = "Bearer " .. config.token,
        ["X-BTR-CLIENT-NAME"] = "XPLANE-12-LUA",
        ["X-BTR-CLIENT-VERSION"] = VERSION,
    }

    local response, status = http_post(API_URL .. "/api/track", body, headers)

    if not response then
        return false, status
    end

    if status ~= 200 and status ~= 201 then
        -- Token might be expired
        if status == 401 then
            state.logged_in = false
            config.token = ""
            save_config()
            return false, "Session expired - please log in again"
        end
        return false, "Track failed (status " .. status .. ")"
    end

    return true, nil
end

-- Logout
local function logout()
    config.token = ""
    state.logged_in = false
    save_config()
end

-- Tracking loop (called every frame)
function tracking_loop()
    if not state.logged_in then return end

    local now = os.clock()
    if now - state.last_send >= TRACK_INTERVAL then
        state.last_send = now
        local ok, err = send_position()
        if not ok and err then
            state.last_error = err
            logMsg("Bushtalk: " .. err)
        else
            state.last_error = ""
        end
    end
end

do_every_frame("tracking_loop()")

-- ImGui window
local wnd = nil

local function create_window()
    if wnd then return end

    wnd = float_wnd_create(300, 250, 1, true)
    float_wnd_set_title(wnd, "Bushtalk Radio")
    float_wnd_set_position(wnd, 100, 100)

    float_wnd_set_imgui_builder(wnd, "build_bushtalk_window")
    float_wnd_set_onclose(wnd, "on_window_close")
end

function on_window_close(wnd_handle)
    wnd = nil
    state.show_window = false
end

function build_bushtalk_window(wnd_handle, x, y)
    imgui.SetCursorPosY(10)
    imgui.SetCursorPosX(10)

    if not http_available then
        imgui.TextUnformatted("Error: LuaSocket not available")
        imgui.TextUnformatted("Please install LuaSocket for FlyWithLua.")
        return
    end

    imgui.TextUnformatted("Bushtalk Radio v" .. VERSION)
    imgui.Separator()

    if not state.logged_in then
        -- Login form
        imgui.TextUnformatted("Username:")
        local changed, new_val = imgui.InputText("##username", input_username, 64)
        if changed then input_username = new_val end

        imgui.TextUnformatted("Password:")
        changed, new_val = imgui.InputText("##password", input_password, 64)
        if changed then input_password = new_val end

        imgui.Spacing()

        if imgui.Button("Login", 280, 30) then
            if input_username ~= "" and input_password ~= "" then
                local ok, err = authenticate(input_username, input_password)
                if not ok then
                    state.last_error = err or "Login failed"
                else
                    input_password = ""
                end
            else
                state.last_error = "Enter username and password"
            end
        end

        if state.last_error ~= "" then
            imgui.Spacing()
            imgui.TextUnformatted("Error: " .. state.last_error)
        end
    else
        -- Logged in view
        imgui.TextUnformatted("Logged in as: " .. config.username)
        imgui.TextUnformatted("Status: Tracking")
        imgui.Spacing()
        imgui.Separator()
        imgui.Spacing()

        -- Position info
        if state.last_position then
            local pos = state.last_position
            imgui.TextUnformatted(string.format("Position: %.4f, %.4f", pos.PLANE_LATITUDE, pos.PLANE_LONGITUDE))
            imgui.TextUnformatted(string.format("Altitude: %.0f ft AGL", pos.ALTITUDE_ABOVE_GROUND))
            imgui.TextUnformatted(string.format("Speed: %.0f kts", pos.GROUND_VELOCITY))
            imgui.TextUnformatted(string.format("Heading: %.0f", pos.MAGNETIC_COMPASS))
            imgui.TextUnformatted("Tail: " .. pos.ATC_ID)
        else
            imgui.TextUnformatted("Waiting for position data...")
        end

        if state.last_error ~= "" then
            imgui.Spacing()
            imgui.TextUnformatted("Error: " .. state.last_error)
        end

        imgui.Spacing()
        imgui.Separator()
        imgui.Spacing()

        if imgui.Button("Logout", 135, 25) then
            logout()
        end
    end
end

-- Menu item to show/hide window (Plugins > FlyWithLua > FlyWithLua Macros)
add_macro("Bushtalk Radio", "create_window()")

-- Auto-show window on load
create_window()

logMsg("Bushtalk Radio v" .. VERSION .. " loaded")
