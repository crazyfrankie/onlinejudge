local key = KEYS[1]
-- 用户输入的 code
local expectedCode = ARGV[1]
local cntKey = key..":cnt"

-- 转成一个数字
local cnt = tonumber(redis.call("get",cntKey))
local code = redis.call("get", key)

if cnt == nil or cnt < 0 then
-- 处理 cnt 为 nil 的情况，例如初始化 cnt 为 0 或者执行某种错误处理逻辑
-- 说明用户一直输错 ，有人搞你
    return -1
elseif expectedCode == code then
-- 输对了
-- 用完了不能再用了
    redis.call("set", cntKey, -1)
    redis.call("expire", cntKey, 600)  -- 重新设置过期时间为 600 秒
    return 0
else
-- 用户手抖输错了
-- 可验证次数减1
    redis.call("decr", cntKey)
    return -2
end