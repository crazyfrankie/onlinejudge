-- 获取限流对象的键
local key = KEYS[1]

-- 窗口大小，表示时间窗口的长度，例如 1000 毫秒
local window = tonumber(ARGV[1])

-- 阈值，表示在窗口时间内允许的最大请求次数
local threshold = tonumber(ARGV[2])

-- 当前时间，单位为毫秒
local now = tonumber(ARGV[3])

-- 计算时间窗口的起始时间，窗口的开始时间 = 当前时间 - 窗口大小
local min = now - window

-- 移除时间窗口之外的请求数据
-- 从有序集合中删除所有 `score` 值在窗口起始时间之前的元素
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)

-- 统计当前时间窗口内的请求数量
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

-- 判断当前请求数是否超过阈值
if cnt >= threshold then
    -- 如果请求数超出阈值，返回 "true"，表示限流
    return "true"
else
    -- 如果请求数未超出阈值，将当前请求时间戳作为 `score` 和 `member` 添加到有序集合中
    redis.call('ZADD', key, now, now)
    -- 返回 "false"，表示未限流
    return "false"
end


---- 1, 2, 3, 4, 5, 6, 7这是你的元素
---- ZREMRANGEBYSCORE key1 0 6
---- 7 执行完之后
--
---- 限流对象
--local key = KEYS[1]
---- 窗口大小
--local window = tonumber(ARGV[1])
---- 阈值
--local threshold = tonumber(ARGV[2])
--local now = tonumber(ARGV[3])
---- 窗口的起始时间
--local min = now - window
--
--redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
--local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')
---- local cnt = redis.call('ZCOUNT', key, min, '+inf')
--if cnt >= threshold then
--    -- 执行限流
--    return "true"
--else
--    -- 把 score 和 member 都设置成 now
--    redis.call('ZADD',key, now, now)
--    redis.call('PREFIX', key, window)
--    return "false"
--end