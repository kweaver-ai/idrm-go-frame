package enum

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type FlowchartClassEnum Object

var (
	FlowchartClassConfigStatusNormal      = New[FlowchartClassEnum](1, "normal", "正常")
	FlowchartClassConfigStatusMissingRole = New[FlowchartClassEnum](2, "missingRole", "缺失角色")
)

type DictFlowchartClassEnum Object

var (
	DictFlowchartClassConfigStatusNormal      = New[DictFlowchartClassEnum](1, "normal", "Dict正常")
	DictFlowchartClassConfigStatusMissingRole = New[DictFlowchartClassEnum](2, "missingRole", "Dict缺失角色")
)

type SystemRoleClassEnum Object

var (
	SystemRoleClassEnumConfigStatusNormal      = New[SystemRoleClassEnum](1, "normals")
	SystemRoleClassEnumConfigStatusMissingRole = New[SystemRoleClassEnum](2, "missingRoles")
)

// DataKind  基础信息分类
type DataKind Object

var (
	DataKindTypeHuman  = New[DataKind](1<<0, "human", "人")
	DataKindTypeLand   = New[DataKind](1<<1, "land", "地")
	DataKindTypeEvent  = New[DataKind](1<<2, "event", "事")
	DataKindTypeObject = New[DataKind](1<<3, "object", "物")
	DataKindTypeOrg    = New[DataKind](1<<4, "org", "组织")
	DataKindTypeOther  = New[DataKind](1<<5, "other", "其他")
)

func TestToInteger(t *testing.T) {
	i := ToInteger[FlowchartClassEnum]("normal").Int()
	assert.Equal(t, i, 1)

	a2 := ToInteger[FlowchartClassEnum]("xxx", 23).Int()
	assert.Equal(t, a2, 23)
}

func TestToString(t *testing.T) {
	s1 := ToString[FlowchartClassEnum](1)
	assert.Equal(t, s1, "normal")

	a := int8(2)
	s2 := ToString[FlowchartClassEnum](a)
	assert.Equal(t, s2, "missingRole")

	a3 := int8(20)
	s3 := ToString[FlowchartClassEnum](a3, "xxx")
	assert.Equal(t, s3, "xxx")
}

func TestIs(t *testing.T) {
	s1 := Is[FlowchartClassEnum](4)
	assert.Equal(t, s1, false)

	s2 := Is[FlowchartClassEnum]("xxx")
	assert.Equal(t, s2, false)
}

func TestDisplay(t *testing.T) {
	assert.Equal(t, FlowchartClassConfigStatusNormal.Display, "正常")
	assert.Equal(t, Get[FlowchartClassEnum]("normal").Display, "正常")
	assert.Equal(t, len(List[FlowchartClassEnum]()), 2)
	assert.Equal(t, Strings[FlowchartClassEnum](String), "normal,missingRole")
	assert.Equal(t, strings.Join(Values("FlowchartClassEnum"), ","), "normal,missingRole")
	assert.True(t, len(Objs("FlowchartClassEnum")) > 0)
}

func TestBits(t *testing.T) {
	s1 := BitsMerge[DataKind]([]string{"human", "land"})
	assert.Equal(t, uint32(3), s1)

	s2 := BitsSplit[DataKind](3)
	assert.Equal(t, 2, len(s2))
	assert.Equal(t, "human", s2[0])
	assert.Equal(t, "land", s2[1])

	s3 := BitsTransfer[DataKind]([]string{"human", "land"})
	assert.Equal(t, 2, len(s3))
	assert.Equal(t, uint32(1), s3[0])
	assert.Equal(t, uint32(2), s3[1])
}

func TestQuery(t *testing.T) {
	assert.NotNil(t, Query("FlowchartClassEnum", "正常"))
	assert.NotNil(t, Query("FlowchartClassEnum", "normal"))
}

func TestObjects(t *testing.T) {
	objs := Objects[DataKind]()
	bs, _ := json.Marshal(objs)
	fmt.Print(string(bs))
}

func TestStringDict(t *testing.T) {
	d := StringDict[FlowchartClassEnum, DictFlowchartClassEnum]("normal")
	assert.Equal(t, true, d.Display == "Dict正常")

	d1 := TypeDict("FlowchartClassEnum", "DictFlowchartClassEnum", "normal")
	assert.Equal(t, true, d1.Display == "Dict正常")
}
