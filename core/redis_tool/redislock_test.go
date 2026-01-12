package redis_tool

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

// 测试锁自动续期
func TestRedisLock(t *testing.T) {
	redisClient := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}
	ctx := context.Background()
	key := "test_key_TestSevAutoRenewSuccess"
	var wg sync.WaitGroup
	wg.Add(2)

	// 线程1
	go func() {
		defer wg.Done()
		lock := NewAutoRedisLock(redisClient.Write, key, "dd")
		//lock.SetExpire(1)
		_, err := lock.LockCtx(ctx)
		if err != nil {
			t.Errorf("Lock() returned unexpected error: %v", err)
			return
		}
		defer lock.UnLockCtx(ctx)
		log.Println("线程1：自旋锁加锁成功")
		time.Sleep(time.Second * 10)
		log.Println("线程1：任务执行结束")
	}()

	//线程2
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 7)
		log.Println("线程2：开始抢夺锁资源")
		lock := NewRedisLock(redisClient.Write, key)
		r, _ := lock.LockCtx(ctx)
		if r == false {
			defer lock.UnLockCtx(ctx)
			t.Errorf("线程2 Lock() not expectation success")
			return
		}
		log.Printf("线程2：获取到资源抢夺锁资源=%v\n", r)
	}()
	wg.Wait()
}

func TestLockFunc(t *testing.T) {
	redisClient := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}
	ctx := context.Background()
	key := "test_key_TestSevAutoRenewSuccessSSSS"
	value := "ddCC"
	var wg sync.WaitGroup
	wg.Add(2)
	// 线程1
	go func() {
		defer wg.Done()
		lock, err := Lock(ctx, key, value)
		if err != nil {
			t.Errorf("Lock() returned unexpected error: %v", err)
			return
		}
		defer UnLockCtx(ctx, key, value)
		log.Printf("线程1：自旋锁加锁成功=%v\n", lock)
		time.Sleep(time.Second * 10)
		log.Println("线程1：任务执行结束")
	}()

	//线程2
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 7)
		log.Println("线程2：开始抢夺锁资源")
		lock, _ := Lock(ctx, key, value)
		if lock == false {
			defer UnLockCtx(ctx, key, value)
			t.Errorf("线程2 Lock() not expectation success")
			return
		}
		log.Printf("线程2：获取到资源抢夺锁资源=%v\n", lock)
	}()
	wg.Wait()
}
