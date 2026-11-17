-- Miru's next-episode button. It draws a small translucent button in the
-- bottom-right corner during the last minute of an episode that has a
-- following entry in mpv's playlist, and advances the playlist when clicked.
-- Everything else (transport controls, seek bar) stays with mpv's own OSC.
--
-- mpv's default left click toggles pause, so the button only takes the click
-- over while the pointer is inside it: a forced binding is added on hover and
-- removed again when the pointer leaves, which leaves every other click with
-- mpv's own meaning.

local mp = require 'mp'

local SHOW_REMAINING = 60
local LABEL = 'Next ▸'
local KEY_NAME = 'miru-next-click'

local overlay = mp.create_osd_overlay('ass-events')
overlay.res_x = 1280
overlay.res_y = 720
overlay.z = 100
overlay.hidden = true

local visible = false
local hovered = false
local bound = false
local layoutWidth, layoutHeight = 0, 0
local button = {x = 0, y = 0, w = 0, h = 0}
local font = 18

-- layout sizes the button in video pixels and keeps a readable minimum, so a
-- small player window still shows a usable target instead of scaling a
-- 1280x720 design down with it.
local function layout()
    local dims = mp.get_property_native('osd-dimensions')
    local width = dims and dims.w or 0
    local height = dims and dims.h or 0
    if width < 1 or height < 1 then
        width, height = 1280, 720
    end
    if width == layoutWidth and height == layoutHeight then
        return false
    end
    layoutWidth, layoutHeight = width, height
    overlay.res_x, overlay.res_y = width, height
    local boxHeight = math.max(34, math.floor(height * 0.075))
    local boxWidth = math.max(132, math.floor(boxHeight * 3.6))
    button.w, button.h = boxWidth, boxHeight
    button.x = width - math.max(16, math.floor(width * 0.03)) - boxWidth
    button.y = height - math.max(24, math.floor(height * 0.07)) - boxHeight
    font = math.max(15, math.floor(boxHeight * 0.44))
    return true
end

local function ass_events()
    local alpha = hovered and '4C' or '73'
    local rect = string.format(
        '{\\an7\\pos(%d,%d)\\bord0\\shad0\\p1\\1c&H1C1C1C&\\1a&H%s&}m 0 0 l %d 0 %d %d 0 %d',
        button.x, button.y, alpha, button.w, button.w, button.h, button.h
    )
    local label = string.format(
        '{\\an5\\pos(%d,%d)\\bord0\\shad0\\fs%d\\1c&HFFFFFF&\\1a&H18&}%s',
        button.x + button.w / 2, button.y + button.h / 2, font, LABEL
    )
    return rect .. '\n' .. label
end

local function draw()
    overlay.hidden = not visible
    if visible then
        overlay.data = ass_events()
        mp.set_property_native('user-data/miru-next-ass', overlay.data)
    end
    overlay:update()
    mp.set_property_native('user-data/miru-next-visible', visible)
    mp.set_property_native('user-data/miru-next-hovered', hovered)
end

local function advance()
    if not visible then
        return
    end
    mp.commandv('playlist-next')
end

local function click()
    if visible and hovered then
        advance()
    end
end

local function bind(enabled)
    if enabled and not bound then
        mp.add_forced_key_binding('MBTN_LEFT', KEY_NAME, click)
        bound = true
    elseif not enabled and bound then
        mp.remove_key_binding(KEY_NAME)
        bound = false
    end
end

local function set_hovered(inside)
    inside = visible and inside
    if inside == hovered then
        return
    end
    hovered = inside
    bind(hovered)
    draw()
end

local function set_visible(show)
    if show == visible then
        return
    end
    visible = show
    if not show then
        hovered = false
        bind(false)
    end
    draw()
end

local function refresh()
    if layout() then
        draw()
    end
    local duration = mp.get_property_number('duration', 0) or 0
    local position = mp.get_property_number('time-pos')
    local index = mp.get_property_number('playlist-pos', -1)
    local count = mp.get_property_number('playlist-count', 0) or 0
    if not position or duration <= 0 or index < 0 then
        return set_visible(false)
    end
    local remaining = duration - position
    local has_next = count > 0 and index < count - 1
    set_visible(has_next and remaining > 0 and remaining <= SHOW_REMAINING)
end

local function pointer_inside(mouse)
    local dims = mp.get_property_native('osd-dimensions')
    if not visible or not mouse or not dims then
        return false
    end
    local mx = mouse.x - (dims.ml or 0)
    local my = mouse.y - (dims.mt or 0)
    return mx >= button.x and mx <= button.x + button.w and my >= button.y and my <= button.y + button.h
end

mp.register_script_message('miru-next-activate', advance)
mp.observe_property('time-pos', 'number', refresh)
mp.observe_property('duration', 'number', refresh)
mp.observe_property('playlist-pos', 'native', refresh)
mp.observe_property('playlist-count', 'native', refresh)
mp.observe_property('osd-dimensions', 'native', refresh)
mp.observe_property('mouse-pos', 'native', function(_, mouse)
    set_hovered(pointer_inside(mouse))
end)
mp.register_event('file-loaded', refresh)
mp.register_event('end-file', function()
    set_hovered(false)
end)

mp.set_property_native('user-data/miru-next-visible', false)
