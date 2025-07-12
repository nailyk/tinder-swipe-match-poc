#!lua name=tinder

redis.register_function("swipeAndCheckMatch", function(keys, args)
    -- keys[1] = "swipes:{A}:{B}" sorted user pair key
    -- args[1] = from user
    -- args[2] = to user
    -- args[3] = swipe action ("like" or "dislike")
    local userPairRedisKey = keys[1]
    local fromUser = args[1]
    local toUser = args[2]
    local swipeAction = args[3]

    redis.call('HSET', userPairRedisKey, fromUser .. '_swipe', swipeAction)

    local reverseSwipeAction = redis.call('HGET', userPairRedisKey, toUser .. '_swipe')

    if swipeAction == "like" and reverseSwipeAction == "like" then
        redis.call('SADD', 'matches:' .. fromUser, toUser)
        redis.call('SADD', 'matches:' .. toUser, fromUser)

        return 1
    end

    return 0
end
)
