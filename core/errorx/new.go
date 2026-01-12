package errorx

import "fmt"

const (
	publicPreCoder = "Basic.Public."
)

var errorCodeMap = make(map[string]ErrorCodeInfo)

type Module struct {
	preCode      string
	errorCodeMap map[string]ErrorCodeInfo
}

func New(preCode string) *Module {
	return &Module{preCode: preCode, errorCodeMap: errorCodeMap}
}

// Solution 该方法还必须在Module方法之后
func (m *Module) Solution(code, desc, cause, solution string) *ErrorCodeInfo {
	e := &ErrorCodeInfo{}
	if m.preCode == "" {
		m.preCode = publicPreCoder
		m.errorCodeMap = errorCodeMap
	}
	e.code = m.preCode + code
	e.description = desc
	e.cause = cause
	e.solution = solution
	//append
	if _, ok := m.errorCodeMap[e.code]; ok {
		panic(fmt.Sprintf("error code is not allowed to repeat, code: %s", e.code))
	}
	m.errorCodeMap[e.code] = *e
	return e
}

// Cause 该方法还必须在Module方法之后
func (m *Module) Cause(code, desc, cause string) *ErrorCodeInfo {
	return m.Solution(code, desc, cause, "")
}

// Description 该方法还必须在Module方法之后
func (m *Module) Description(code, desc string) *ErrorCodeInfo {
	return m.Solution(code, desc, "", "")
}
