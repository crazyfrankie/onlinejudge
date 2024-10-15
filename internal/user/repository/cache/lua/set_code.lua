-- 验证码在 Redis 上的 key
-- phone_code:login:131xxxxxxx
local key = KEYS[1]
-- 使用次数，也就是验证次数
-- phone_code:login:131xxxxxxx:cnt
local cntKey = key..":cnt"

-- 预期中的验证码
local val = ARGV[1]

-- 过期时间
-- 验证码的有效时间是十分钟，600秒
local ttl = tonumber(redis.call("ttl",key))

-- -1 是 key 存在但是没有过期时间
if ttl == -1 then
--  有人误操作导致 key 冲突
    return -2
elseif ttl == -2 or ttl < 540 then
--  后续如果验证码有不同的过期时间，要在这里优化
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", key, 600)
    return 0
else
--  已经发送了一个验证码，但是还不到一分钟
    return -1
end

