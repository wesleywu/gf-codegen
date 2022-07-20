package common

import (
	"context"
	_ "embed"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gview"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

var (
	ColumnTypeStr       = []string{"char", "varchar", "varchar2", "tinytext", "text", "mediumtext", "longtext", "binary", "varbinary", "blob"}
	ColumnTypeDate      = []string{"date"}
	ColumnTypeTime      = []string{"datetime", "time", "timestamp"}
	ColumnTypeNumber    = []string{"tinyint", "smallint", "mediumint", "int", "integer", "bigint", "float", "double", "decimal", "numeric", "bit"}
	ColumnNameNotEdit   = []string{"created_by", "created_at", "updated_by", "updated_at", "deleted_at"}
	ColumnNameNotList   = []string{"updated_by", "updated_at", "deleted_at"}
	ColumnNameNotDetail = []string{"updated_by", "updated_at", "deleted_at"}
	ColumnNameNotQuery  = []string{"updated_by", "updated_at", "deleted_at", "remark"}
)

// IsExistInArray 判断 value 是否存在在切片array中
func IsExistInArray(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

// GetColumnLength 获取字段长度
func GetColumnLength(columnType string) int {
	start := strings.Index(columnType, "(")
	end := strings.Index(columnType, ")")
	result := ""
	if start >= 0 && end >= 0 {
		result = columnType[start+1 : end-1]
	}
	return gconv.Int(result)
}

// IsStringObject 判断是否是数据库字符串类型
func IsStringObject(dataType string) bool {
	return IsExistInArray(dataType, ColumnTypeStr)
}

// IsDateObject 判断是否是数据库时间类型
func IsDateObject(dataType string) bool {
	return IsExistInArray(dataType, ColumnTypeDate)
}

// IsTimeObject 判断是否是数据库时间类型
func IsTimeObject(dataType string) bool {
	return IsExistInArray(dataType, ColumnTypeTime)
}

// IsNumberObject 是否数字类型
func IsNumberObject(dataType string) bool {
	return IsExistInArray(dataType, ColumnTypeNumber)
}

func GetGoModuleName() (string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return "", gerror.New("获取本地路径失败")
	}
	goModuleFile := path.Join(curDir, "go.mod")
	if !gfile.Exists(goModuleFile) {
		return "", gerror.New("请在项目根路径下运行本程序，并要求根路径存在go.mod")
	}
	return GetGoModule(goModuleFile)
}

func GetGoModule(file string) (string, error) {
	goModContent, err := ioutil.ReadFile(file)
	if err != nil {
		return "", gerror.Newf("获取%s内容失败", file)
	}
	lines := gstr.Split(string(goModContent), "\n")

	reGoModule, _ := regexp.Compile("module\\s+([^\\s]+)")
	for _, line := range lines {
		matches := reGoModule.FindStringSubmatch(line)
		if len(matches) >= 1 {
			return matches[1], nil
		}
	}
	return "", gerror.Newf("%s文件中找不到module指令", file)
}

func getBusinessName(tableName string, tablePrefix []string) string {
	return removeTablePrefix(tableName, tablePrefix)
}

//删除表前缀
func removeTablePrefix(tableName string, tablePrefix []string) string {
	if !g.IsEmpty(tablePrefix) {
		for _, str := range tablePrefix {
			if strings.HasPrefix(tableName, str) {
				return strings.Replace(tableName, str, "", 1) //注意，只替换一次
			}
		}
	}
	return tableName
}

func LoadTableDefYaml(ctx context.Context, tableName string, yamlInputPath string, goModuleName string, cache map[string]*TableDef) (*TableDef, error) {
	cached, found := cache[tableName]
	if found {
		return cached, nil
	}
	def, err := loadCodeDefYaml(ctx, tableName, yamlInputPath)
	if err != nil {
		return nil, err
	}
	table := &TableDef{}
	err = gconv.Struct(def.Table, table)
	if err != nil {
		return nil, err
	}
	table.SetVariableNames(goModuleName)
	table.ColumnMap = def.Columns
	table.VirtualColumnMap = def.VirtualColumns

	table.Columns = columnsSlice(def.Columns, false)
	table.VirtualColumns = columnsSlice(def.VirtualColumns, true)
	table.AddColumns = addColumnsSlice(def.AddColumns)
	table.EditColumns = editColumnsSlice(def.EditColumns)
	table.ListColumns = listColumnsSlice(def.ListColumns)
	table.QueryColumns = queryColumnsSlice(def.QueryColumns)
	table.DetailColumns = detailColumnsSlice(def.DetailColumns)

	createdAt, hasCreatedAt := table.ColumnMap["created_at"]
	if hasCreatedAt {
		table.CreatedAtColumn = createdAt
	}
	createBy, hasCreatedBy := table.ColumnMap["created_by"]
	table.HasCreatedBy = hasCreatedBy
	if hasCreatedBy {
		table.CreatedByColumn = createBy
	}
	_, hasUpdateBy := table.ColumnMap["updated_by"]
	table.HasUpdatedBy = hasUpdateBy
	table.RefColumns = gmap.NewListMap()
	table.VirtualQueryRelated = make(map[string]*TableDef)

	err = table.ProcessColumns(ctx, yamlInputPath, goModuleName, cache)
	if err != nil {
		return nil, err
	}
	cache[tableName] = table
	return table, nil
}

func loadCodeDefYaml(ctx context.Context, tableName string, yamlInputPath string) (*CodeGenDef, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, gerror.New("获取本地路径失败")
	}

	yamlFile := path.Join(curDir, yamlInputPath, tableName+".yaml")
	bytes, err1 := ioutil.ReadFile(yamlFile)
	if err1 != nil {
		return nil, gerror.New("读取 " + yamlFile + " 失败")
	}

	var def = &CodeGenDef{}
	err = yaml.Unmarshal(bytes, def)
	return def, err
}

func TemplateEngine() *gview.View {
	view := gview.New()
	_ = view.SetConfigWithMap(g.Map{
		"Delimiters": []string{"{{", "}}"},
	})
	view.BindFuncMap(g.Map{
		"UcFirst": func(str string) string {
			return gstr.UcFirst(str)
		},
		"Sum": func(a, b int) int {
			return a + b
		},
		"CaseCamelLower": gstr.CaseCamelLower, //首字母小写驼峰
		"CaseCamel":      gstr.CaseCamel,      //首字母大写驼峰
		"CaseKebab":      gstr.CaseKebab,      //全小写短横线分隔
		"HasSuffix":      gstr.HasSuffix,      //是否存在后缀
		"ContainsI":      gstr.ContainsI,      //是否包含子字符串
		"VueTag": func(t string) string {
			return t
		},
		"IsEmpty": g.IsEmpty, //是否为空
		"IsNotEmpty": func(value interface{}) bool {
			return !g.IsEmpty(value)
		},
	})
	return view
}

func TrimBreak(str string) (rStr string, err error) {
	var b []byte
	rStr = str
	// 将连续多个换行换为一个
	if b, err = gregex.Replace("(([ \t]*)\r?\n){2,}", []byte("\n"), []byte(str)); err != nil {
		return
	}
	// 去掉行尾空格
	if b, err = gregex.Replace("(?:(?:[ \t]+)\r?\n)", []byte("\n"), b); err != nil {
		return
	}
	// 在函数结尾和 import 结尾增加一个空行
	if b, err = gregex.Replace("(?:\n(////|[)}])\n)", []byte("\n$1\n\n"), b); err != nil {
		return
	}
	// 在函数结尾和 import 结尾增加一个空行
	if b, err = gregex.Replace("(?:\n(import(?:[ \t]+)?[(])\n)", []byte("\n\n$1\n"), b); err != nil {
		return
	}
	rStr = gconv.String(b)
	return
}

func WriteFile(fileName, data string, cover bool) (err error) {
	if !gfile.Exists(fileName) || cover {
		var f *os.File
		f, err = gfile.Create(fileName)
		if err == nil {
			_, _ = f.WriteString(data)
		}
		_ = f.Close()
	}
	return
}

func GetDataType(sqlType string) (dataType string, isUnsigned bool) {
	t, _ := gregex.ReplaceString(`\(.+\)`, "", sqlType)
	typeSlice := gstr.Split(gstr.Trim(t), " ")
	if len(typeSlice) > 1 {
		isUnsigned = gstr.ToLower(typeSlice[1]) == "unsigned"
	}
	dataType = gstr.ToLower(typeSlice[0])
	return
}

func columnsSlice(columnMap map[string]*ColumnDef, isVirtual bool) []*ColumnDef {
	columns := make([]*ColumnDef, len(columnMap))
	i := 0
	for name, column := range columnMap {
		column.Name = name
		column.IsVirtual = isVirtual
		columns[i] = column
		i++
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Sort < columns[j].Sort
	})
	return columns
}

func listColumnsSlice(columnMap map[string]*ListColumnDef) []*ListColumnDef {
	columns := make([]*ListColumnDef, len(columnMap))
	i := 0
	for name, column := range columnMap {
		column.Name = name
		columns[i] = column
		i++
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Sort < columns[j].Sort
	})
	return columns
}

func addColumnsSlice(columnMap map[string]*AddColumnDef) []*AddColumnDef {
	columns := make([]*AddColumnDef, len(columnMap))
	i := 0
	for name, column := range columnMap {
		column.Name = name
		columns[i] = column
		i++
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Sort < columns[j].Sort
	})
	return columns
}

func editColumnsSlice(columnMap map[string]*EditColumnDef) []*EditColumnDef {
	columns := make([]*EditColumnDef, len(columnMap))
	i := 0
	for name, column := range columnMap {
		column.Name = name
		columns[i] = column
		i++
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Sort < columns[j].Sort
	})
	return columns
}

func queryColumnsSlice(columnMap map[string]*QueryColumnDef) []*QueryColumnDef {
	columns := make([]*QueryColumnDef, len(columnMap))
	i := 0
	for name, column := range columnMap {
		column.Name = name
		columns[i] = column
		i++
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Sort < columns[j].Sort
	})
	return columns
}

func detailColumnsSlice(columnMap map[string]*DetailColumnDef) []*DetailColumnDef {
	columns := make([]*DetailColumnDef, len(columnMap))
	i := 0
	for name, column := range columnMap {
		column.Name = name
		columns[i] = column
		i++
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Sort < columns[j].Sort
	})
	return columns
}
