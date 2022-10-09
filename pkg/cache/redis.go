package cache

import (
	"bbs/pkg/logging"
	"bbs/pkg/util"
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

var redisClients = make(map[string]*Redis)
var ctx = context.Background()

type Redis struct {
	client        *redis.Client
	clusterClient *redis.ClusterClient
	trace         *logging.Cache
}

const (
	DefaultRedisClient = "default-cache-client"
	MinIdleConns       = 50
	PoolSize           = 20
	MaxRetries         = 3
)

func setDefaultOptions(opt *redis.Options) {
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 2 * time.Second
	}

	if opt.ReadTimeout == 0 {
		//默认值为3秒
		opt.ReadTimeout = 2 * time.Second
	}

	if opt.ReadTimeout == 0 {
		//默认值与ReadTimeout相等
		opt.ReadTimeout = 2 * time.Second
	}

	if opt.PoolTimeout == 0 {
		//默认为ReadTimeout + 1秒（4s）
		opt.PoolTimeout = 10 * time.Second
	}
	if opt.IdleTimeout == 0 {
		//默认值为5秒
		opt.IdleTimeout = 10 * time.Second
	}
}

func setDefaultClusterOptions(opt *redis.ClusterOptions) {
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 2 * time.Second
	}

	if opt.ReadTimeout == 0 {
		//默认值为3秒
		opt.ReadTimeout = 2 * time.Second
	}

	if opt.ReadTimeout == 0 {
		//默认值与ReadTimeout相等
		opt.ReadTimeout = 2 * time.Second
	}

	if opt.PoolTimeout == 0 {
		//默认为ReadTimeout + 1秒（4s）
		opt.PoolTimeout = 10 * time.Second
	}
	if opt.IdleTimeout == 0 {
		//默认值为5秒
		opt.IdleTimeout = 10 * time.Second
	}
}

func InitRedis(clientName string, opt *redis.Options, trace *logging.Cache) error {
	if len(clientName) == 0 {
		return errors.New("empty client name")
	}

	if len(opt.Addr) == 0 {
		return errors.New("empty addr")
	}

	setDefaultOptions(opt)
	client := redis.NewClient(opt)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return errors.Wrap(err, "ping cache err addr : "+opt.Addr)
	}
	InitLua(client)
	redisClients[clientName] = &Redis{
		client: client,
		trace:  trace,
	}
	return nil
}

func InitClusterRedis(clientName string, opt *redis.ClusterOptions, trace *logging.Cache) error {
	if len(clientName) == 0 {
		return errors.New("empty client name")
	}
	if len(opt.Addrs) == 0 {
		return errors.New("empty addrs")
	}
	setDefaultClusterOptions(opt)
	//NewClusterClient执行过程中会连接redis集群并, 并尝试发送("cluster", "info")指令去进行多次连接,
	//如果这里传入很多连接地址，并且连接地址都不可用的情况下会阻塞很长时间
	client := redis.NewClusterClient(opt)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("ping cache err  addrs : %v", opt.Addrs))
	}
	redisClients[clientName] = &Redis{
		clusterClient: client,
	}
	return nil
}

func GetRedisClient(name string) *Redis {
	if client, ok := redisClients[name]; ok {
		return client
	}
	return nil
}

func GetRedisClusterClient(name string) *Redis {
	if client, ok := redisClients[name]; ok {
		return client
	}
	return nil
}

// Set set some <key,value> into cache
func (r *Redis) Set(key string, value interface{}, ttl time.Duration) error {
	if len(key) == 0 {
		return errors.New("empty key")
	}
	defer func() {
		trace(key, value, "set", r)
	}()

	if r.client != nil {
		if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
			return errors.Wrapf(err, "cache set key: %s err", key)
		}
		return nil
	}

	//集群版
	if err := r.clusterClient.Set(ctx, key, value, ttl).Err(); err != nil {
		return errors.Wrapf(err, "cache set key: %s err", key)
	}
	return nil
}

// Get get some key from cache
func (r *Redis) Get(key string) interface{} {
	if len(key) == 0 {
		CacheStdLogger.Println("empty key")
		return nil
	}
	defer func() {
		trace(key, "", "get", r)
	}()

	if r.client != nil {
		value, err := r.client.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			CacheStdLogger.Printf("cache get key: %s err %v", key, err)
		}
		return value
	}

	value, err := r.clusterClient.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		CacheStdLogger.Printf("cache get key: %s err %v", key, err)
	}
	return value
}

func (r *Redis) HGet(key, field string) (string, error) {
	if len(key) == 0 {
		CacheStdLogger.Println("empty key")
		return "", errors.New("HGet empty key ")
	}
	defer func() {
		trace(key, "", "get", r)
	}()

	if r.client != nil {
		value, err := r.client.HGet(ctx, key, field).Result()
		if err != nil && err != redis.Nil {
			CacheStdLogger.Printf("cache get key: %s field: %s err %v", key, field, err)
		}
		return value, err
	}

	value, err := r.clusterClient.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		CacheStdLogger.Printf("cache get key: %s err %v", key, err)
	}
	return value, err
}

func (r *Redis) HSet(key, field string, value interface{}, ttl time.Duration) error {

	if len(key) == 0 {
		return errors.New("empty key")
	}
	defer func() {
		trace(key, value, "hset", r)
	}()

	if r.client != nil {
		if err := r.client.HSet(ctx, key, field, value, ttl).Err(); err != nil {
			return errors.Wrapf(err, "cache set key: %s err", key)
		}
		return nil
	}

	//集群版
	if err := r.clusterClient.HSet(ctx, key, field, value, ttl).Err(); err != nil {
		return errors.Wrapf(err, "cache set key: %s err", key)
	}
	return nil
}

func (r *Redis) HMSet(key string, value []interface{}, ttl time.Duration) error {
	if len(key) == 0 {
		return errors.New("empty key")
	}

	if r.client != nil {
		if err := r.client.HMSet(ctx, key, value...).Err(); err != nil {
			return errors.Wrapf(err, "cache set key: %s err", key)
		}
		if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
			return errors.Wrapf(err, "expire key: %s", key)
		}
		return nil
	}

	//集群版
	if err := r.clusterClient.HMSet(ctx, key, value...).Err(); err != nil {
		return errors.Wrapf(err, "cache set key: %s err", key)
	}
	if err := r.clusterClient.Expire(ctx, key, ttl).Err(); err != nil {
		return errors.Wrapf(err, "expire key: %s", key)
	}
	return nil
}
func (r *Redis) EvalSha(sha, opt string, key []string, args ...interface{}) (interface{}, error) {
	funcName := "cache.EvalSha"
	if r.client != nil {
		value, err := r.client.EvalSha(ctx, sha, key, args).Result()
		if err != nil {
			CacheStdLogger.Printf("%s opt: %s ,err: %v", funcName, opt, err)
		}
		return value, errors.WithMessage(err, funcName)
	}
	value, err := r.clusterClient.EvalSha(ctx, sha, key, args).Result()
	if err != nil && err != redis.Nil {
		CacheStdLogger.Printf("%s opt: %s ,err: %v", funcName, opt, err)
	}
	return value, err
}

func (r *Redis) GetStr(key string) (value string, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	defer func() {
		trace(key, value, "get", r)
	}()

	if r.client != nil {
		value, err = r.client.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return "", errors.Wrapf(err, "cache get key: %s err", key)
		}
		return
	}

	value, err = r.clusterClient.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return "", errors.Wrapf(err, "cache get key: %s err", key)
	}
	return
}

// TTL get some key from cache
func (r *Redis) TTL(key string) (time.Duration, error) {
	if len(key) == 0 {
		return 0, errors.New("empty key")
	}
	if r.client != nil {
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return -1, errors.Wrapf(err, "cache get key: %s err", key)
		}
		return ttl, nil
	}
	ttl, err := r.clusterClient.TTL(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return -1, errors.Wrapf(err, "cache get key: %s err", key)
	}

	return ttl, nil
}

// Expire expire some key
func (r *Redis) Expire(key string, ttl time.Duration) (bool, error) {
	if len(key) == 0 {
		return false, errors.New("empty key")
	}
	if r.client != nil {
		ok, err := r.client.Expire(ctx, key, ttl).Result()
		return ok, err
	}
	ok, err := r.clusterClient.Expire(ctx, key, ttl).Result()
	return ok, err
}

// ExpireAt expire some key at some time
func (r *Redis) ExpireAt(key string, ttl time.Time) (bool, error) {
	if len(key) == 0 {
		return false, errors.New("empty key")
	}
	if r.client != nil {
		ok, err := r.client.ExpireAt(ctx, key, ttl).Result()
		return ok, err
	}
	ok, err := r.clusterClient.ExpireAt(ctx, key, ttl).Result()
	return ok, err

}

func (r *Redis) Exists(keys ...string) (bool, error) {
	if len(keys) == 0 {
		return false, errors.New("empty keys")
	}
	if r.client != nil {
		value, err := r.client.Exists(ctx, keys...).Result()
		return value > 0, err
	}
	value, err := r.clusterClient.Exists(ctx, keys...).Result()
	return value > 0, err
}

func (r *Redis) IsExist(key string) bool {
	if len(key) == 0 {
		return false
	}
	if r.client != nil {
		value, err := r.client.Exists(ctx, key).Result()
		if err != nil && err != redis.Nil {
			CacheStdLogger.Printf("cmd : Exists ; key : %s ; err : %v", key, err)
		}
		return value > 0
	}
	value, err := r.clusterClient.Exists(ctx, key).Result()
	if err != nil && err != redis.Nil {
		CacheStdLogger.Printf("cmd : Exists ; key : %s ; err : %v", key, err)
	}
	return value > 0
}

func (r *Redis) Delete(key string) error {
	if len(key) == 0 {
		return errors.New("empty key")
	}
	var value int64
	var err error
	defer func() {
		trace(key, value, "del", r)
	}()

	if r.client != nil {
		_, err = r.client.Del(ctx, key).Result()
		return err
	}

	//集群版
	_, err = r.clusterClient.Del(ctx, key).Result()
	return err
}

func (r *Redis) Incr(key string) (value int64, err error) {
	if len(key) == 0 {
		return 0, errors.New("empty key")
	}

	defer func() {
		trace(key, value, "Incr", r)
	}()
	if r.client != nil {
		value, err = r.client.Incr(ctx, key).Result()
		return
	}
	value, err = r.clusterClient.Incr(ctx, key).Result()
	return
}

// Close close cache client
func (r *Redis) Close() error {
	return r.client.Close()
}

// Version cache server version
func (r *Redis) Version() string {
	if r.client != nil {
		server := r.client.Info(ctx, "server").Val()
		spl1 := strings.Split(server, "# Server")
		spl2 := strings.Split(spl1[1], "redis_version:")
		spl3 := strings.Split(spl2[1], "redis_git_sha1:")
		return spl3[0]
	}
	server := r.clusterClient.Info(ctx, "server").Val()
	spl1 := strings.Split(server, "# Server")
	spl2 := strings.Split(spl1[1], "redis_version:")
	spl3 := strings.Split(spl2[1], "redis_git_sha1:")
	return spl3[0]

}

func traceInt(key string, value int64, cmd string, r *Redis) {
	ts := time.Now()
	if r.trace == nil || r.trace.Logger == nil {
		return
	}
	costMillisecond := time.Since(ts).Milliseconds()

	if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
		return
	}
	r.trace.TraceTime = util.CSTLayoutString()
	r.trace.CMD = cmd
	r.trace.Key = key
	r.trace.Value = strconv.FormatInt(value, 10)
	r.trace.CostMillisecond = costMillisecond
	r.trace.Logger.Warn("cache-trace", zap.Any("", r.trace))
}

func trace(key string, value interface{}, cmd string, r *Redis) {
	ts := time.Now()
	if r.trace == nil || r.trace.Logger == nil {
		return
	}
	costMillisecond := time.Since(ts).Milliseconds()

	if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
		return
	}
	r.trace.TraceTime = util.CSTLayoutString()
	r.trace.CMD = cmd
	r.trace.Key = key
	r.trace.Value = value
	r.trace.CostMillisecond = costMillisecond
	r.trace.Logger.Warn("cache-trace", zap.Any("", r.trace))
}
