# 枚举
Golang自带枚举，但是这些枚举还是和当前的接口不适用，所以有了该枚举包

## 适用场景
当前的项目，一般是前端和后端交互使用的是英文单词表示枚举, 但是实际后端在数据库保存的是整形，这就涉及到一种，字符串值和整型值之间的转换

## 例子
枚举代码
```go
// CommonStatus 项目，任务等状态
type CommonStatus enum.Object

var (
	CommonStatusReady     = enum.New[CommonStatus](1, "ready", "未开始")
	CommonStatusOngoing   = enum.New[CommonStatus](2, "ongoing", "进行中")
	CommonStatusCompleted = enum.New[CommonStatus](3, "completed", "已完成")
)
```
转换代码如下
```go
//从整形的枚举值转换成字符串的枚举值
taskStatus := enum.ToString[constant.CommonStatus](task.Status)
//从字符串的枚举值转换成整形的枚举值
task.Status = enum.ToInteger[constant.CommonStatus](taskStatus).Int8()
```

## 详细使用
### 定义枚举类
对于某一类的枚举，使用单独的类型定义
```go
// CommonStatus 项目，任务等状态
type CommonStatus enum.Object

var (
	CommonStatusReady     = enum.New[CommonStatus](1, "ready", "未开始")
	CommonStatusOngoing   = enum.New[CommonStatus](2, "ongoing", "进行中")
	CommonStatusCompleted = enum.New[CommonStatus](3, "completed", "已完成")
)
```
上述代码，以 `CommonStatusReady`为例
- `1` 是数据保存的值
- `ready`是前端传过来的值
- `未开始`是前端显示的值

