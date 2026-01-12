package enum

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

// IntegerType 整形数据 int，为什么这么大，因为有时候，项目会使用位计算
// int8很明显不够用，所以使用int, 一般int是32位
type IntegerType int

func (ip IntegerType) Int8() int8   { return int8(ip) }
func (ip IntegerType) Int16() int16 { return int16(ip) }
func (ip IntegerType) Int32() int32 { return int32(ip) }
func (ip IntegerType) Int() int     { return int(ip) }

//下面几种方法不支持有符号数据，有符号数据会转成无符号的数据类型

func (ip IntegerType) Uint8() uint8   { return uint8(ip) }
func (ip IntegerType) Uint16() uint16 { return uint16(ip) }
func (ip IntegerType) Uint32() uint32 { return uint32(ip) }
func (ip IntegerType) Uint() uint     { return uint(ip) }

type recordPair struct {
	//stringEnumRecord  string to int reflection
	stringEnumRecord map[string]Object
	//intEnumRecord  int to string reflection
	intEnumRecord map[IntegerType]Object
	//displayEnumRecord  display to int
	displayEnumRecord map[string]Object
	//values
	values []string
}

// recordPairs stringEnumRecord and intEnumRecord pair, key is struct name
type recordPairs map[any]recordPair

// allRecords  record all the EnumRecord  pair
var allRecords recordPairs = make(map[any]recordPair)

// typeRecords 通过字符串记录下实际的枚举类型
var typeRecords = make(map[string]any)

// Object one enum object
type Object struct {
	index   int          // 插入的顺序
	Integer *IntegerType // 枚举整形值
	String  string       // 枚举的字符串表示值，英文
	Display string       // 枚举的前端展示值，中文
}

// enumProperty enum object record
type enumProperty interface {
	string | ~uint | ~uint32 | ~uint16 | ~uint8 | ~int | ~int32 | ~int16 | ~int8
}

// enumInteger enum integer type
type enumInteger interface {
	~uint | ~uint32 | ~uint16 | ~uint8 | ~int | ~int32 | ~int16 | ~int8
}

type PropertyKind Object

var (
	Integer = New[PropertyKind](1, "integer")
	String  = New[PropertyKind](2, "string")
	Display = New[PropertyKind](3, "display")
)

func ObjectName[T any]() string {
	k := fmt.Sprintf("%T", new(T))
	return k[strings.LastIndexAny(k, ".")+1:]
}

// set record the enum
func set[T any](e *T) {
	key := *new(T)
	maps, ok := allRecords[key]
	if !ok {
		maps = recordPair{
			stringEnumRecord:  make(map[string]Object),
			intEnumRecord:     make(map[IntegerType]Object),
			displayEnumRecord: make(map[string]Object),
			values:            make([]string, 0),
		}
		typeRecords[ObjectName[T]()] = key
	}
	a := (*Object)(unsafe.Pointer(e))
	a.index = len(maps.stringEnumRecord)
	maps.stringEnumRecord[a.String] = *a
	maps.intEnumRecord[*a.Integer] = *a
	if a.Display != "" {
		maps.displayEnumRecord[a.Display] = *a
	}
	//record the value
	maps.values = append(maps.values, a.String)
	allRecords[key] = maps
}

// ToString  find the string value of 'i', panic if not the valid enum type or not a valid value, return the defaultValue[0] if not find
func ToString[T any, V enumInteger](i V, defaultValue ...string) string {
	maps, ok := allRecords[*new(T)]
	if !ok {
		log.Panicf("invalid enum struct %v", reflect.TypeOf(new(T)))
	}
	iv := IntegerType(i)
	obj, ok := maps.intEnumRecord[iv]
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return obj.String
}

// ToInteger find the uint8 value of 's' panic if not the valid enum type, return the defaultValue[0] if not find
func ToInteger[T any](s string, defaultValue ...IntegerType) IntegerType {
	maps, ok := allRecords[*new(T)]
	if !ok {
		log.Panicf("invalid enum struct %v", reflect.TypeOf(new(T)))
	}
	var obj Object
	//string
	obj, ok = maps.stringEnumRecord[s]
	if ok {
		return *(obj.Integer)
	}
	//display
	obj, ok = maps.displayEnumRecord[s]
	if ok {
		return *(obj.Integer)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// New generate EnumClass
func New[T any](i IntegerType, s ...string) T {
	if len(s) <= 0 {
		log.Panicf("invalid enum object %v, missing string value", reflect.TypeOf(new(T)))
	}
	e := Object{
		Integer: &i,
		String:  s[0],
	}
	if len(s) >= 2 {
		e.Display = s[1]
	}
	p := (*T)(unsafe.Pointer(&e))
	set[T](p)
	return *p
}

// Is  check whether value v is a valid enum value
func Is[T any, P enumProperty](v P) bool {
	maps, ok := allRecords[*new(T)]
	if !ok {
		log.Panicf("invalid enum struct %v", reflect.TypeOf(new(T)))
	}
	var ev interface{}
	ev = v
	switch ev.(type) {
	case string:
		_, ok := maps.stringEnumRecord[ev.(string)]
		return ok
	default:
		pv, _ := strconv.Atoi(fmt.Sprintf("%v", ev))
		_, ok := maps.intEnumRecord[(IntegerType)(pv)]
		return ok
	}
	return false
}

// Get  获取枚举值常量
func Get[T any, P enumProperty](v P) (o *T) {
	destObj := GetObj[T](v)
	enumObj := new(T)
	obj := (*Object)(unsafe.Pointer(enumObj))
	obj.Display = destObj.Display
	obj.index = destObj.index
	obj.Integer = destObj.Integer
	obj.String = destObj.String
	return enumObj
}

// GetObj  获取枚举值常量
func GetObj[T any, P enumProperty](v P) *Object {
	maps, ok := allRecords[*new(T)]
	if !ok {
		log.Panicf("invalid enum struct %v", reflect.TypeOf(new(T)))
	}
	//新建返回结果
	var ev interface{} = v
	var sourceObj Object
	switch ev.(type) {
	case string:
		key := fmt.Sprintf("%v", ev)
		sourceObj = maps.stringEnumRecord[key]
	default:
		pv, _ := strconv.Atoi(fmt.Sprintf("%v", ev))
		sourceObj = maps.intEnumRecord[IntegerType(pv)]
	}
	return &sourceObj
}

func convert[T any](obj Object) *T {
	enumObj := new(T)
	destObj := (*Object)(unsafe.Pointer(enumObj))
	destObj.String = obj.String
	destObj.Integer = obj.Integer
	destObj.Display = obj.Display
	return enumObj
}

func List[T any]() []T {
	maps, ok := allRecords[*new(T)]
	if !ok {
		log.Panicf("invalid enum struct %v", reflect.TypeOf(new(T)))
	}
	results := make([]T, 0, len(maps.stringEnumRecord))
	for _, obj := range maps.stringEnumRecord {
		results = append(results, *convert[T](obj))
	}
	return results
}

func Objects[T any]() []Object {
	maps, ok := allRecords[*new(T)]
	if !ok {
		log.Panicf("invalid enum struct %v", reflect.TypeOf(new(T)))
	}
	results := make([]Object, 0, len(maps.stringEnumRecord))
	for _, obj := range maps.stringEnumRecord {
		results = append(results, obj)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})
	return results
}

func Strings[T any](p PropertyKind) string {
	vs := Objects[T]()
	values := make([]string, 0)
	for _, item := range vs {
		switch p.String {
		case Integer.String:
			values = append(values, fmt.Sprintf("%v", item.Integer))
		case String.String:
			values = append(values, item.String)
		case Display.String:
			values = append(values, item.Display)
		}
	}
	return strings.Join(values, ",")
}

// Values 枚举对象的  String  slice，方便参数检查
func Values(name string) []string {
	key := typeRecords[name]
	maps, ok := allRecords[key]
	if !ok {
		log.Panicf("invalid enum struct %v", name)
	}
	return maps.values
}

func Query[P enumProperty](name string, key P) *Object {
	typeKey, has := typeRecords[name]
	if !has {
		return nil
	}
	maps, ok := allRecords[typeKey]
	if !ok {
		log.Panicf("invalid enum name %v", name)
	}
	var obj Object
	keyStr := fmt.Sprintf("%v", key)
	//string
	obj, ok = maps.stringEnumRecord[keyStr]
	if ok {
		return &obj
	}
	//display
	obj, ok = maps.displayEnumRecord[keyStr]
	if ok {
		return &obj
	}
	//integer
	pv, _ := strconv.Atoi(fmt.Sprintf("%v", keyStr))
	obj, ok = maps.intEnumRecord[IntegerType(pv)]
	if ok {
		return &obj
	}
	return nil
}

// Objs 枚举对象的  String  slice，方便参数检查
func Objs(name string) []Object {
	objs := make([]Object, 0)

	key := typeRecords[name]
	maps, ok := allRecords[key]
	if !ok {
		return objs
	}
	for _, obj := range maps.stringEnumRecord {
		objs = append(objs, obj)
	}
	return objs
}

func BitsMerge[T any](ss []string) uint32 {
	if len(ss) <= 0 {
		return 0
	}
	var res uint32 = 0
	for _, s := range ss {
		res += ToInteger[T](s).Uint32()
	}
	return res
}

func BitsTransfer[T any](vs []string) []uint32 {
	res := make([]uint32, 0)
	for _, s := range vs {
		res = append(res, ToInteger[T](s).Uint32())
	}
	return res
}

func BitsSplit[T any](d uint32) []string {
	dataKinds := make([]string, 0)
	for i := 0; i <= 31; i++ {
		dk := uint32(1) << i
		if Is[T](dk) && dk&d > 0 {
			dataKinds = append(dataKinds, ToString[T](dk))
		}
	}
	return dataKinds
}

// StringDict 从S到D的枚举映射查询，当前映射的关键是枚举属性String
func StringDict[S any, D any](str string) *D {
	source := GetObj[S](str)
	if source == nil {
		log.Panicf("invalid source enum  %v", str)
	}
	return Get[D](source.String)
}

// TypeDict 从source到dest的枚举映射查询，当前映射的关键是枚举属性是string
func TypeDict[P enumProperty](s, d string, key P) *Object {
	source := Query(s, key)
	if source == nil {
		log.Panicf("invalid source enum  %v", s)
	}
	return Query(d, source.String)
}
