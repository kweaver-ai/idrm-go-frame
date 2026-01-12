package redis_tool

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/idx"

	"github.com/go-redis/redis/v8"
)

const (
	randomLen       = 16
	tolerance       = 500 // milliseconds
	millisPerSecond = 1000
)

var (
	luaLock = `if redis.call("GET", KEYS[1]) == ARGV[1] then
		redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
		return "OK"
		else
		return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
		end`
	luaUnLock = `if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
		else
		return 0
		end`
)

// A RedisLock is a redis lock.
type RedisLock struct {
	key      string        // 锁key
	value    string        // 锁的值，随机数
	seconds  uint32        // 锁过期时间,单位秒，防止死锁
	client   redis.Cmdable // 锁客户端，暂时只有redis
	unlockCh chan struct{} // 解锁通知通道
}

// NewRedisLock returns a RedisLock.
func NewRedisLock(client redis.Cmdable, key string) *RedisLock {
	d := &RedisLock{
		client: client,
		key:    key,
		value:  idx.NewUUID().String(),
		//isAutoRenew: false,
	}
	d.unlockCh = make(chan struct{}, 0)
	return d
}

func NewAutoRedisLock(client redis.Cmdable, key string, value string) *RedisLock {
	d := &RedisLock{
		client: client,
		key:    key,
		value:  value,
	}
	d.unlockCh = make(chan struct{}, 0)
	return d
}

// 重入性key、value需要一样,需要手工处理
func Lock(ctx context.Context, key, value string) (bool, error) {
	return LockCtx(ctx, key, value, tolerance)
}

func LockCtx(ctx context.Context, key, value string, seconds int) (bool, error) {
	resp, err := rl.Write.Eval(ctx, luaLock, []string{key}, value, seconds).Result()
	if err == redis.Nil || resp == nil {
		log.Printf("error on acquiring lock err == redis.Nil || resp == nil for %s, %s", key, err.Error())
		return false, err
	} else if err != nil {
		log.Printf("error on acquiring err != nil lock for %s, %s", key, err.Error())
		return false, err
	}
	reply, ok := resp.(string)
	if ok && reply == "OK" {
		return true, nil
	}
	log.Printf("unknown reply when acquiring lock for %s: %v", key, resp)
	return false, err
}

func UnLockCtx(ctx context.Context, key, value string) (bool, error) {
	resp, err := rl.Write.Eval(ctx, luaUnLock, []string{key}, []string{value}).Result()
	if err != nil {
		return false, err
	}
	reply, ok := resp.(int64)
	if !ok {
		return false, nil
	}
	return reply == 1, nil
}

// 天生具备重入性
func (rl *RedisLock) Lock() (bool, error) {
	return rl.LockCtx(context.Background())
}

func (rl *RedisLock) LockCtx(ctx context.Context) (bool, error) {
	// 默认锁过期时间为500ms，防止死锁
	expireSeconds, seconds := int(atomic.LoadUint32(&rl.seconds))*millisPerSecond, tolerance
	if expireSeconds > 0 {
		seconds = expireSeconds
	}
	resp, err := rl.client.Eval(ctx, luaLock, []string{rl.key}, rl.value, seconds).Result()
	if err == redis.Nil || resp == nil {
		log.Printf("error on acquiring lock err == redis.Nil || resp == nil for %s, %s", rl.key, err.Error())
		return false, err
	} else if err != nil {
		log.Printf("error on acquiring err != nil  lock for %s, %s", rl.key, err.Error())
		return false, err
	}
	reply, ok := resp.(string)
	if ok && reply == "OK" {
		//if rl.isAutoRenew {
		//	go rl.autoLock(seconds, ctx)
		//}
		return true, nil
	}
	log.Printf("unknown reply when acquiring lock for %s: %v", rl.key, resp)
	return false, nil
}

// UnLock 解锁
func (rl *RedisLock) UnLock() (bool, error) {
	return rl.UnLockCtx(context.Background())
}

func (rl *RedisLock) UnLockCtx(ctx context.Context) (bool, error) {
	resp, err := rl.client.Eval(ctx, luaUnLock, []string{rl.key}, []string{rl.value}).Result()
	if err != nil {
		return false, err
	}
	reply, ok := resp.(int64)
	if !ok {
		return false, nil
	}
	close(rl.unlockCh)
	return reply == 1, nil
}

// SetExpire sets the expiration.
func (rl *RedisLock) SetExpire(seconds int) {
	atomic.StoreUint32(&rl.seconds, uint32(seconds))
}

// 锁自动续期
func (rl *RedisLock) autoLock(seconds int, ctx context.Context) {
	// 创建一个定时器NewTicker, 每过期时间的3分之2触发一次
	loopTime := time.Duration(seconds*2/3) * millisPerSecond
	ticker := time.NewTicker(loopTime)
	defer ticker.Stop()
	//确认锁与锁续期打包原子化
	for {
		select {
		case <-rl.unlockCh:
			return
		case <-ticker.C:
			_, err := rl.LockCtx(ctx)
			if err != nil {
				log.Println("autoLock failed:", err)
				return
			}
		}
	}
}
