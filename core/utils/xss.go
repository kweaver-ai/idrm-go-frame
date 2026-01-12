package utils

import "strings"

func XssEscape(values string) string {
	if values == "" {
		return values
	}
	special := strings.NewReplacer(`<`, `&lt;`, `>`, `&gt;`, `select`, `查询`, `drop`, `删除表`, `delete`, `删除数据`, `update`, `更新`, `insert`,
		`插入`, `SELECT`, `查询`, `DROP`, `删除表`, `DELETE`, `删除数据`, `UPDATE`, `更新`, `INSERT`, `插入`, `script`, `脚本`, `SCRIPT`, `脚本`, `ALTER`,
		`修改结构`, `alter`, `修改结构`, `create`, `创建`, `CREATE`, `创建`)
	return special.Replace(values)
}
