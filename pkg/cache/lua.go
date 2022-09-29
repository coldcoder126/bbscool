package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v9"
)

var Lua = make(map[int]string)

const (
	PointSub = iota
	PostToSet
	PostToQue
)

var scripts = map[int]string{
	PointSub:  "local subNum = tonumber(ARGV[1]) \nlocal curNum = tonumber(cache.call('hget',KEYS[1],KEYS[2])) \nif (subNum > curNum) \nthen \nreturn {-1} \nelse \ncache.call('hincrby',KEYS[1],KEYS[2],-subNum) \nreturn {1} end",
	PostToSet: "",
	PostToQue: "",
}

//将帖子信息放入zset中
//local val = ARGV[1]
//local len = tonumber(ARGV[2])
//redis.call('ZADD',KEYS[1],tonumber(os.time()),val)
//redis.call('ZREMRANGEBYRANK',KEYS[1],len,-1)
//redis.call('EXPIREAT',KEYS[1],tonumber(os.time()))

//将帖子信息按时间顺序放入队列中
//key = KEYS[1]
//val = ARGV[1]
//local len = tonumber(ARGV[2])
//redis.call('LPUSH',key,val)
//redis.call('LTRIM',key,0,len)
//redis.call('EXPIREAT',KEYS[1],tonumber(os.time()))

func InitLua(client *redis.Client) error {
	// 清除Redis上所有脚本
	fmt.Println("清除Redis上的所有脚本")
	ctx = context.Background()
	err := client.ScriptFlush(ctx).Err()
	if err != nil {
		return errors.New("flush redis script error")
	}

	for k, v := range scripts {
		sha, err := client.ScriptLoad(ctx, v).Result()
		if err != nil {
			return errors.New("load redis script error")
		}
		Lua[k] = sha
	}
	return nil
}
