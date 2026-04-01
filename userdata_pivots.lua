-- Aseprite script - https://www.aseprite.org/docs/scripting/
--
-- 1. Open the .ase files in the .img folder.
-- 2. In the pivots layer, click on the cel and move it using the Move Tool.
-- 3. Then, when you run this script, the calculated pivot (offset) will be written to the Tag's Userdata field.
--
-- See parseUserDataOffsets() in utils.go 

local sprite = app.activeSprite
local pivotLayer = nil

for _, layer in ipairs(sprite.layers) do
    if layer.name == "pivots" then
        pivotLayer = layer
        break
    end
end

for _, tag in ipairs(sprite.tags) do
    local pivotLayer = pivotLayer:cel(tag.fromFrame)
    local x = pivotLayer.bounds.x + 8
    local y = pivotLayer.bounds.y + 8
    tag.data = x .. "," .. y
end
