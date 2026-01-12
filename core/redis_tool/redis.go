package redis_tool

import (
	"context"
	"log"
	"time"
)

/*------------------------------------ 字符 操作 ------------------------------------*/

// Set 设置 key的值
func Set(ctx context.Context, key string, value interface{}) bool {
	result, err := rl.Write.Set(ctx, key, value, 0).Result()
	if err != nil {
		log.Print(err)
		return false
	}
	return result == "OK"
}

// SetEX 设置 key的值并指定过期时间
func SetEX(ctx context.Context, key string, value interface{}, ex time.Duration) bool {
	result, err := rl.Write.Set(ctx, key, value, ex).Result()
	if err != nil {
		log.Print(err)
		return false
	}
	return result == "OK"
}

// Get 获取 key的值
func Get(ctx context.Context, key string) (bool, string) {
	result, err := rl.Read.Get(ctx, key).Result()
	if err != nil {
		log.Print(err)
		return false, ""
	}
	return true, result
}

// GetSet 设置新值获取旧值
func GetSet(ctx context.Context, key string, value interface{}) (bool, string) {
	oldValue, err := rl.Write.GetSet(ctx, key, value).Result()
	if err != nil {
		log.Print(err)
		return false, ""
	}
	return true, oldValue
}

// Incr key值每次加一 并返回新值
func Incr(ctx context.Context, key string) int64 {
	val, err := rl.Write.Incr(ctx, key).Result()
	if err != nil {
		log.Print(err)
	}
	return val
}

// IncrBy key值每次加指定数值 并返回新值
func IncrBy(ctx context.Context, key string, incr int64) int64 {
	val, err := rl.Write.IncrBy(ctx, key, incr).Result()
	if err != nil {
		log.Print(err)
	}
	return val
}

// IncrByFloat key值每次加指定浮点型数值 并返回新值
func IncrByFloat(ctx context.Context, key string, incrFloat float64) float64 {
	val, err := rl.Write.IncrByFloat(ctx, key, incrFloat).Result()
	if err != nil {
		log.Print(err)
	}
	return val
}

// Decr key值每次递减 1 并返回新值
func Decr(ctx context.Context, key string) int64 {
	val, err := rl.Write.Decr(ctx, key).Result()
	if err != nil {
		log.Print(err)
	}
	return val
}

// DecrBy key值每次递减指定数值 并返回新值
func DecrBy(ctx context.Context, key string, incr int64) int64 {
	val, err := rl.Write.DecrBy(ctx, key, incr).Result()
	if err != nil {
		log.Print(err)
	}
	return val
}

// Del 删除 key
func Del(ctx context.Context, key string) bool {
	result, err := rl.Write.Del(ctx, key).Result()
	if err != nil {
		return false
	}
	return result == 1
}

// Expire 设置 key的过期时间
func Expire(ctx context.Context, key string, ex time.Duration) bool {
	result, err := rl.Write.Expire(ctx, key, ex).Result()
	if err != nil {
		return false
	}
	return result
}

/*------------------------------------ list 操作 ------------------------------------*/

// LPush 从列表左边插入数据，并返回列表长度
func LPush(ctx context.Context, key string, values ...interface{}) int64 {
	result, err := rl.Write.LPush(ctx, key, values).Result()
	if err != nil {
		log.Print(err)
	}
	return result
}

// RPush 从列表右边插入数据，并返回列表长度
func RPush(ctx context.Context, key string, values ...interface{}) int64 {
	result, err := rl.Write.RPush(ctx, key, values).Result()
	if err != nil {
		log.Print(err)
	}
	return result
}

// LPop 从列表左边删除第一个数据，并返回删除的数据
func LPop(ctx context.Context, key string) (bool, string) {
	val, err := rl.Write.LPop(ctx, key).Result()
	if err != nil {
		log.Print(err)
		return false, ""
	}
	return true, val
}

// RPop 从列表右边删除第一个数据，并返回删除的数据
func RPop(ctx context.Context, key string) (bool, string) {
	val, err := rl.Write.RPop(ctx, key).Result()
	if err != nil {
		log.Print(err)
		return false, ""
	}
	return true, val
}

// LIndex 根据索引坐标，查询列表中的数据
func LIndex(ctx context.Context, key string, index int64) (bool, string) {
	val, err := rl.Read.LIndex(ctx, key, index).Result()
	if err != nil {
		log.Print(err)
		return false, ""
	}
	return true, val
}

// LLen 返回列表长度
func LLen(ctx context.Context, key string) int64 {
	val, err := rl.Read.LLen(ctx, key).Result()
	if err != nil {
		log.Print(err)
	}
	return val
}

// LRange 返回列表的一个范围内的数据，也可以返回全部数据
func LRange(ctx context.Context, key string, start, stop int64) []string {
	vales, err := rl.Read.LRange(ctx, key, start, stop).Result()
	if err != nil {
		log.Print(err)
	}
	return vales
}

// LRem 从列表左边开始，删除元素data， 如果出现重复元素，仅删除 count次
func LRem(ctx context.Context, key string, count int64, value interface{}) bool {
	_, err := rl.Write.LRem(ctx, key, count, value).Result()
	if err != nil {
		log.Print(err)
	}
	return true
}

// LInsert 在列表中 pivot 元素的后面插入 value
func LInsert(ctx context.Context, key string, pivot int64, value interface{}) bool {
	err := rl.Write.LInsert(ctx, key, "after", pivot, value).Err()
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}

/*------------------------------------ set 操作 ------------------------------------*/

// SAdd 添加元素到集合中
func SAdd(ctx context.Context, key string, values ...interface{}) bool {
	err := rl.Write.SAdd(ctx, key, values).Err()
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}

// SCard 获取集合元素个数
func SCard(ctx context.Context, key string) int64 {
	size, err := rl.Read.SCard(ctx, key).Result()
	if err != nil {
		log.Print(err)
	}
	return size
}

// SIsMember 判断元素是否在集合中
func SIsMember(ctx context.Context, key string, value interface{}) bool {
	ok, err := rl.Read.SIsMember(ctx, key, value).Result()
	if err != nil {
		log.Print(err)
	}
	return ok
}

// SMembers 获取集合所有元素
func SMembers(ctx context.Context, key string) []string {
	es, err := rl.Read.SMembers(ctx, key).Result()
	if err != nil {
		log.Print(err)
	}
	return es
}

// SRem 删除 key集合中的 data元素
func SRem(ctx context.Context, key string, values ...interface{}) bool {
	_, err := rl.Write.SRem(ctx, key, values).Result()
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}

// SPopN 随机返回集合中的 count个元素，并且删除这些元素
func SPopN(ctx context.Context, key string, count int64) []string {
	vales, err := rl.Write.SPopN(ctx, key, count).Result()
	if err != nil {
		log.Print(err)
	}
	return vales
}

/*------------------------------------ hash 操作 ------------------------------------*/

// HSet 根据 key和 field字段设置，field字段的值
func HSet(ctx context.Context, key, field, value string) bool {
	err := rl.Write.HSet(ctx, key, field, value).Err()
	if err != nil {
		return false
	}
	return true
}

// HGet 根据 key和 field字段，查询field字段的值
func HGet(ctx context.Context, key, field string) string {
	val, err := rl.Read.HGet(ctx, key, field).Result()
	if err != nil {
		log.Print(err)
	}
	return val
}

// HMGet 根据key和多个字段名，批量查询多个 hash字段值
func HMGet(ctx context.Context, key string, fields ...string) []interface{} {
	vales, err := rl.Read.HMGet(ctx, key, fields...).Result()
	if err != nil {
		panic(err)
	}
	return vales
}

// HGetAll 根据 key查询所有字段和值
func HGetAll(ctx context.Context, key string) map[string]string {
	data, err := rl.Read.HGetAll(ctx, key).Result()
	if err != nil {
		log.Print(err)
	}
	return data
}

// HKeys 根据 key返回所有字段名
func HKeys(ctx context.Context, key string) []string {
	fields, err := rl.Read.HKeys(ctx, key).Result()
	if err != nil {
		log.Print(err)
	}
	return fields
}

// HLen 根据 key，查询hash的字段数量
func HLen(ctx context.Context, key string) int64 {
	size, err := rl.Read.HLen(ctx, key).Result()
	if err != nil {
		log.Print(err)
	}
	return size
}

// HMSet 根据 key和多个字段名和字段值，批量设置 hash字段值
func HMSet(ctx context.Context, key string, value map[string]interface{}) bool {
	result, err := rl.Write.HMSet(ctx, key, value).Result()
	if err != nil {
		log.Print(err)
		return false
	}
	return result
}

// HSetNX 如果 field字段不存在，则设置 hash字段值
func HSetNX(ctx context.Context, key, field string, value interface{}) bool {
	result, err := rl.Write.HSetNX(ctx, key, field, value).Result()
	if err != nil {
		log.Print(err)
		return false
	}
	return result
}

// HDel 根据 key和字段名，删除 hash字段，支持批量删除
func HDel(ctx context.Context, key string, fields ...string) bool {
	_, err := rl.Write.HDel(ctx, key, fields...).Result()
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}

// HExists 检测 hash字段名是否存在
func HExists(ctx context.Context, key, field string) bool {
	result, err := rl.Read.HExists(ctx, key, field).Result()
	if err != nil {
		log.Print(err)
		return false
	}
	return result
}
