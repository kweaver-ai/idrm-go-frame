package redis_tool

import (
	"context"
	"fmt"
	"testing"
)

func getRedisClient() *Redis {
	client := InitRedisConf(RedisConf{
		Addrs: []string{"10.4.71.138:6379"},
		Pass:  "",
		//Port:     6379,
		Type:     standaloneType,
		PoolSize: 10,
	})
	return client
}

func TestSetKey(t *testing.T) {
	getRedisClient()
	b := Set(context.Background(), "sssss", "SSAQ")
	fmt.Printf("==设置key结果====%v=", b)
	c, res := Get(context.Background(), "sssss")
	fmt.Printf("==获取设置key结果====%v==%s=", c, res)
}

func TestHSetKey(t *testing.T) {
	getRedisClient()
	b1 := HSet(context.Background(), "ssssa", "SSAQ", "SSAAA1")
	fmt.Println("==设置hkey结果==b1==", b1)
	b2 := HSet(context.Background(), "ssssa", "SS2AQ", "SSAAA2")
	fmt.Println("==设置hkey结果==b2===", b2)
	b3 := HSet(context.Background(), "ssssa", "SS3AQ", "SSAAA3")
	fmt.Println("==设置hkey结果==b3==", b3)
	res := HGet(context.Background(), "ssssa", "SSAQ")
	fmt.Println("==获取设置hkey结果====", res)
	res2 := HMGet(context.Background(), "ssssa", "SSAQ", "SS3AQ")
	fmt.Println("==获取设置hkey结果=多个值==", res2)
	res3 := HGetAll(context.Background(), "ssssa")
	fmt.Println("==获取设置hkey结果=map==", res3)
	for k, v := range res3 {
		fmt.Println(k, v)
	}
}
