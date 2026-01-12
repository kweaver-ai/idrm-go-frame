package models

import (
	"database/sql/driver"
	"strconv"

	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agcodes"
	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agerrors"
)

type ModelID string

func NewModelID(id uint64) ModelID {
	return ModelID(strconv.FormatUint(id, 10))
}

func (m ModelID) Uint64() uint64 {
	if len(m) == 0 {
		return 0
	}

	uintId, err := strconv.ParseUint(string(m), 10, 64)
	if err != nil {
		coder := agcodes.New("Public.InvalidParameter", "参数值异常", "", "ID需要修改为可解析为数字的字符串", err, "")
		panic(agerrors.NewCode(coder))
	}

	return uintId
}

// Value 实现数据库驱动所支持的值
// 没有该方法会将ModelID在驱动层转换后string，导致与数据库定义类型不匹配
func (m ModelID) Value() (driver.Value, error) {
	return m.Uint64(), nil
}
