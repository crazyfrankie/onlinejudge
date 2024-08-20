local key = KEYS[1]
-- 用户输入的 code
local expectedCode = ARGV[1]
local cntKey = key..":cnt"

-- 转成一个数字
local cnt = tonumber(redis.call("get",cntKey))

if cnt < 0 then
-- 说明用户一直输错 ，有人搞你
    return -1
elseif expectedCode == code then
-- 输对了
-- 用完了不能再用了
    redis.call("set", cntKey, -1)
    return 0
else
-- 用户手抖输错了
-- 可验证次数减1
    redis.call("decr", cntKey, -1)
    return -2
end

