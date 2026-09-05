package cache

const slidingReadSrc = `
local key = KEYS[1]
local ttl = tonumber(ARGV[1])
local cmd = ARGV[2]

local args = { key }
for i = 3, #ARGV do
  args[#args + 1] = ARGV[i]
end

local value = redis.call(cmd, unpack(args))

if value ~= nil and value ~= false then
  redis.call("EXPIRE", key, ttl)
end

return value
`
