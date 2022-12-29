package common

import (
	"context"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
)

const (
	RelatedTablePrefix     = "Rltd"
	RelatedTableJsonPrefix = "rltd"
)

type CodeGenDef struct {
	ApiVersion     string                      `yaml:"apiVersion"`     // 代码生成版本，当前为 v1
	Table          *TableDef                   `yaml:"table"`          // 数据库表基本属性
	Columns        map[string]*ColumnDef       `yaml:"columns"`        // 数据库表所有字段
	VirtualColumns map[string]*ColumnDef       `yaml:"virtualColumns"` // 虚拟字段，必须关联到关联表中的字段，通常用于列表、详情和查询
	ListColumns    map[string]*ListColumnDef   `yaml:"listColumns"`    // 列表界面中展示字段
	AddColumns     map[string]*AddColumnDef    `yaml:"addColumns"`     // 新增界面可输入字段
	EditColumns    map[string]*EditColumnDef   `yaml:"editColumns"`    // 编辑界面可输入字段
	QueryColumns   map[string]*QueryColumnDef  `yaml:"queryColumns"`   // 列表界面中可查询字段
	DetailColumns  map[string]*DetailColumnDef `yaml:"detailColumns"`  // 详情界面中展示字段
}

type TableDef struct { // 表属性
	Name                 string                `yaml:"name"`                       // 表名
	Comment              string                `yaml:"comment,omitempty"`          // 表描述
	BackendPackage       string                `yaml:"backendPackage,omitempty"`   // Go文件根目录，通常以 cartx/app/ 打头，下面可以有子目录，对应老方法的 PackageName
	ClassName            string                `yaml:"-"`                          // 对应Go单例名，Name去掉前缀然后转大驼峰
	StructName           string                `yaml:"-"`                          // 对应Go的 struct 名，Name去掉前缀然后转小驼峰
	GoFileName           string                `yaml:"-"`                          // 对应Go的文件名，Name去掉前缀然后转小写下划线分隔
	RouteChildPath       string                `yaml:"-"`                          // 对应的 http route 子路径，小写短横线分隔(Kebab命名规则)
	FrontendModule       string                `yaml:"frontendModule,omitempty"`   // 前端模块路径，对应老方法的 ModuleName
	FrontendPath         string                `yaml:"-"`                          // 前端模块路径，将 FrontendModule 做 Kebab 处理
	FrontendFileName     string                `yaml:"-"`                          // 前端API文件名，Name去掉前缀然后转小写短横线分隔(Kebab命名规则)
	TemplateCategory     string                `yaml:"templateCategory,omitempty"` // 代码生成类型 crud/tree
	PackageName          string                `yaml:"-"`                          // Go 文件的 package
	PackageNameProto     string                `yaml:"-"`                          // proto 文件的 package，由 BackendPackage 将 / 换成 . ，并且全小写，与Java client互通时这就是interface的 package name
	ModuleName           string                `yaml:"-"`                          // Go模块及前端模块名，已废弃
	BusinessName         string                `yaml:"businessName,omitempty"`     // 业务名，如不填写，则由表名去掉前缀得到
	FunctionName         string                `yaml:"functionName,omitempty"`     // 功能名称（用于菜单显示和代码注释）
	FunctionAuthor       string                `yaml:"functionAuthor,omitempty"`   // 功能作者
	TreeCode             string                `yaml:"treeCode,omitempty"`         // tree类型对应的当前记录键字段
	TreeParentCode       string                `yaml:"treeParentCode,omitempty"`   // tree类型对应的父记录查询字段
	TreeName             string                `yaml:"treeName,omitempty"`         // tree类型对应的当前记录显示字段
	Overwrite            bool                  `yaml:"overwrite,omitempty"`        // 生成时是否覆盖现有代码和菜单设置
	SortColumn           string                `yaml:"sortColumn,omitempty"`       // 排序字段
	SortType             string                `yaml:"sortType,omitempty"`         // 排序方式 asc/desc
	ShowDetail           bool                  `yaml:"showDetail,omitempty"`       // 是否有显示详情功能
	IsRpc                bool                  `yaml:"isRpc,omitempty"`            // 是否生成dubbogo rpc代码
	SeparatePackage      bool                  `yaml:"separatePackage,omitempty"`  // 是否将代码生成到单独的目录下
	RpcPort              int                   `yaml:"rpcPort"`                    // rpc provider 服务侦听端口
	CreateTime           *gtime.Time           `yaml:"createTime,omitempty"`       // 当前配置初始生成时间
	UpdateTime           *gtime.Time           `yaml:"updateTime,omitempty"`       // 当前配置最后修改时间
	Id                   int64                 `yaml:"-"`                          // 仅用于迁移 tools_gen_table 时使用
	HasTimeColumnInMain  bool                  `yaml:"-"`                          // 主表字段中是否有时间字段
	HasTimeColumn        bool                  `yaml:"-"`                          // 主表+外表+关联表中是否有时间字段被用到
	HasCheckboxColumn    bool                  `yaml:"-"`                          // 主表中是否有html类型为checkbox的字段
	HasUpFileColumn      bool                  `yaml:"-"`                          // 主表+外表+关联表中是否有UpFile字段
	HasConversion        bool                  `yaml:"-"`                          // 是否需要字段值转换
	CreatedAtColumn      *ColumnDef            `yaml:"-"`                          // created_at字段
	CreatedByColumn      *ColumnDef            `yaml:"-"`                          // created_by字段
	HasCreatedBy         bool                  `yaml:"-"`                          // 是否有created_by字段
	HasUpdatedBy         bool                  `yaml:"-"`                          // 是否有updated_by字段
	IsPkInEdit           bool                  `yaml:"-"`                          // 主键是否出现在 EditColumn 中
	PkColumns            map[string]*ColumnDef `yaml:"-"`                          // 主键列信息（可以有多个）
	ColumnMap            map[string]*ColumnDef `yaml:"-"`                          // 所有列的map，key为 Name
	Columns              []*ColumnDef          `yaml:"-"`                          // 数据库表所有字段
	VirtualColumnMap     map[string]*ColumnDef `yaml:"-"`                          // 所有虚拟列的map，key为 Name
	VirtualColumns       []*ColumnDef          `yaml:"-"`                          // 所有虚拟字段
	ListColumns          []*ListColumnDef      `yaml:"-"`                          // 列表界面中展示字段
	AddColumns           []*AddColumnDef       `yaml:"-"`                          // 新增界面可输入字段
	EditColumns          []*EditColumnDef      `yaml:"-"`                          // 编辑界面可输入字段
	QueryColumns         []*QueryColumnDef     `yaml:"-"`                          // 列表界面中可查询字段
	DetailColumns        []*DetailColumnDef    `yaml:"-"`                          // 详情界面中展示字段
	OrmWithMapping       string                `yaml:"-"`                          // orm with 映射信息
	RefColumns           *gmap.ListMap         `yaml:"-"`                          // 作为关联表时，要被查询的所有数据列信息
	RelatedTableMap      *gmap.ListMap         `yaml:"-"`                          // 关联表map
	RelatedTables        []interface{}         `yaml:"-"`                          // 关联表slice
	ClassNameWhenRelated string                `yaml:"-"`                          // 当作为 relatedTable 时的类名
	JsonNameWhenRelated  string                `yaml:"-"`                          // 当作为 relatedTable 时的json名
	CombinedClassName    string                `yaml:"-"`                          // 如果为二级嵌套，同ClassName；如果为三级嵌套，则为外表Class+关联表Class
	HasVirtualQueries    bool                  `yaml:"-"`                          // 是否有虚拟字段参与查询
	VirtualQueryRelated  map[string]*TableDef  `yaml:"-"`                          // 虚拟字段参与查询的关联表
	FkColumnNameSet      *gset.StrSet          `yaml:"-"`                          // 所有的外键字段
	FkColumnsNotInList   []*ColumnDef          `yaml:"-"`                          // 没有出现在 list 列表中的 ForeignKeyColumnName 字段
	AllRelatedTableMap   *gmap.ListMap         `yaml:"-"`                          // 所有的被关联表map，包含二级嵌套和三级嵌套
	AllRelatedTables     []interface{}         `yaml:"-"`                          // 所有的被关联表slice，包含二级嵌套和三级嵌套
}

type ColumnDef struct { // 字段基本属性
	Name                   string                `yaml:"-"`                                // 字段名
	Comment                string                `yaml:"comment,omitempty"`                // 字段描述
	SqlType                string                `yaml:"sqlType,omitempty"`                // 字段数据类型
	Sort                   int                   `yaml:"sort"`                             // 显示排序
	GoType                 string                `yaml:"goType,omitempty"`                 // go字段类型，可以不填（会根据ColumnType自动判断）
	ProtoType              string                `yaml:"-"`                                // protobuf类型
	ConvertFunc            string                `yaml:"-"`                                // 对该类型的类型转换函数
	GoField                string                `yaml:"goField,omitempty"`                // go字段变量名，可以不填（会根据ColumnName按驼峰规则自动填充）
	HtmlField              string                `yaml:"htmlField,omitempty"`              // 字段前端变量名，可以不填（会根据ColumnName按小驼峰规则自动填充）
	HtmlType               string                `yaml:"htmlType,omitempty"`               // 前端控件类型
	IsPk                   bool                  `yaml:"isPk,omitempty"`                   // 是否为主键（目前仅支持单字段主键，不支持联合主键）
	IsIncrement            bool                  `yaml:"isIncrement,omitempty"`            // 是否为自增长字段
	IsRequired             bool                  `yaml:"isRequired,omitempty"`             // 是否必填
	DictType               string                `yaml:"dictType,omitempty"`               // 参照的字典名称
	RelatedTableName       string                `yaml:"relatedTableName,omitempty"`       // 关联表名称
	RelatedKeyColumn       map[string]*ColumnDef `yaml:"-"`                                // 关联表的主键
	RelatedValueColumnName string                `yaml:"relatedValueColumnName,omitempty"` // 关联表Value字段名
	IsCascade              bool                  `yaml:"isCascade,omitempty"`              // 是否需要级联查询（需要与关联表联合使用，级联规则为 当前表.ParentColumnName = 级联表.CascadeColumnName）
	ParentColumnName       string                `yaml:"parentColumnName,omitempty"`       // 级联查询时本表中的上级字段名
	CascadeColumnName      string                `yaml:"cascadeColumnName,omitempty"`      // 级联查询时关联表中对应字段名
	IsCascadeParent        bool                  `yaml:"-"`                                // 是否为级联查询的上级字段
	CascadeParent          *ColumnDef            `yaml:"-"`                                // 级联父字段指针
	CascadeChildrenColumns *gset.StrSet          `yaml:"-"`                                // 所有级联子字段名（按级联顺序）
	IsVirtual              bool                  `yaml:"-"`                                // 是否虚拟字段，如果是虚拟，必须给出 ForeignXXX 三个字段的正确值
	ForeignTableName       string                `yaml:"foreignTableName,omitempty"`       // 虚拟字段实际所在的表
	ForeignKeyColumnName   string                `yaml:"foreignKeyColumnName,omitempty"`   // 与虚拟字段所在表的主键关联（参照）之当前表字段，即外键。注意，当前表中不应当出现多个字段同时关联某一个表的主键
	ForeignValueColumnName string                `yaml:"foreignValueColumnName,omitempty"` // 虚拟字段对应所在表的实际字段
	ForeignTableClass      string                `yaml:"-"`                                // 虚拟字段所在表的ClassName
	CombinedTableClass     string                `yaml:"-"`                                // 关联、虚拟值字段所属实际表的ClassName
	CombinedHtmlTableClass string                `yaml:"-"`                                // 关联、虚拟值字段所属实际表的前端类名（用于构建字典填充和下拉框内容延迟填充）
	CombinedHtmlField      string                `yaml:"-"`                                // 关联、虚拟字段的前端变量名
}

type ListColumnDef struct {
	Name              string     `yaml:"-"`                           // 字段名
	Sort              int        `yaml:"sort"`                        // 排序
	HtmlType          string     `yaml:"htmlType,omitempty"`          // 前端控件类型
	IsInlineEditable  bool       `yaml:"isInlineEditable,omitempty"`  // 是否允许行内编辑（目前仅应用于 yes/no 及 正常/停用 字典字段）
	MinWidth          int        `yaml:"minWidth,omitempty"`          // 列最小显示宽度
	IsFixed           bool       `yaml:"isFixed,omitempty"`           // 在列表中是否固定在最左边
	IsOverflowTooltip bool       `yaml:"isOverflowTooltip,omitempty"` // 在列表中是否省略一行显示不下的内容并将完整内容放在 tooltip 中
	Base              *ColumnDef `yaml:"-"`                           // 对应字段
	Comment           string     `yaml:"-"`                           // 字段描述
	GoType            string     `yaml:"-"`                           // go字段类型，可以不填（会根据ColumnType自动判断）
	GoField           string     `yaml:"-"`                           // go字段变量名，可以不填（会根据ColumnName按驼峰规则自动填充）
	HtmlField         string     `yaml:"-"`                           // 字段前端变量名，可以不填（会根据ColumnName按小驼峰规则自动填充）
}

type AddColumnDef struct {
	Name      string     `yaml:"-"`                  // 字段名
	Sort      int        `yaml:"sort"`               // 排序
	HtmlType  string     `yaml:"htmlType,omitempty"` // 前端控件类型
	Base      *ColumnDef `yaml:"-"`                  // 对应字段
	Comment   string     `yaml:"-"`                  // 字段描述（从字段基本属性中复制）
	GoType    string     `yaml:"-"`                  // go字段类型（从字段基本属性中复制）
	GoField   string     `yaml:"-"`                  // go字段变量名（从字段基本属性中复制）
	HtmlField string     `yaml:"-"`                  // 字段前端变量名（从字段基本属性中复制）
}

type EditColumnDef struct {
	Name       string     `yaml:"-"`                    // 字段名
	Sort       int        `yaml:"sort"`                 // 排序
	HtmlType   string     `yaml:"htmlType,omitempty"`   // 前端控件类型
	IsDisabled bool       `yaml:"isDisabled,omitempty"` // 是否为不可编辑状态
	Base       *ColumnDef `yaml:"-"`                    // 对应字段
	Comment    string     `yaml:"-"`                    // 字段描述（从字段基本属性中复制）
	GoType     string     `yaml:"-"`                    // go字段类型（从字段基本属性中复制）
	GoField    string     `yaml:"-"`                    // go字段变量名（从字段基本属性中复制）
	HtmlField  string     `yaml:"-"`                    // 字段前端变量名（从字段基本属性中复制）
}

type QueryColumnDef struct {
	Name            string     `yaml:"-"`                   // 字段名
	Sort            int        `yaml:"sort"`                // 排序
	HtmlType        string     `yaml:"htmlType,omitempty"`  // 前端控件类型
	QueryType       string     `yaml:"queryType,omitempty"` // 查询类型 EQ|LIKE|BETWEEN
	FieldValidation string     `yaml:"-"`                   // 查询请求中的参数验证规则
	FieldConversion string     `yaml:"-"`                   // 查询请求中的必要类型转换
	Base            *ColumnDef `yaml:"-"`                   // 对应字段
	Comment         string     `yaml:"-"`                   // 字段描述（从字段基本属性中复制）
	GoType          string     `yaml:"-"`                   // go字段类型（从字段基本属性中复制）
	GoField         string     `yaml:"-"`                   // go字段变量名（从字段基本属性中复制）
	HtmlField       string     `yaml:"-"`                   // 字段前端变量名（从字段基本属性中复制）
}

type DetailColumnDef struct {
	Name       string     `yaml:"-"`                    // 字段名
	Sort       int        `yaml:"sort"`                 // 排序
	HtmlType   string     `yaml:"htmlType,omitempty"`   // 前端控件类型
	ColSpan    int        `yaml:"colSpan,omitempty"`    // 占据的栏位数（缺省为12，一行总栏位为24，即一行放两个字段的详情）
	IsRowStart bool       `yaml:"isRowStart,omitempty"` // 是否另起新行
	Base       *ColumnDef `yaml:"-"`                    // 对应字段
	Comment    string     `yaml:"-"`                    // 字段描述（从字段基本属性中复制）
	GoType     string     `yaml:"-"`                    // go字段类型（从字段基本属性中复制）
	GoField    string     `yaml:"-"`                    // go字段变量名（从字段基本属性中复制）
	HtmlField  string     `yaml:"-"`                    // 字段前端变量名（从字段基本属性中复制）
}

func (s *TableDef) SetVariableNames(goModuleName string) {
	s.BackendPackage = gstr.TrimLeftStr(s.BackendPackage, "/")
	s.BackendPackage = gstr.TrimRightStr(s.BackendPackage, "/")

	if gstr.Pos(s.BackendPackage, goModuleName) != 0 {
		s.BackendPackage = goModuleName + "/" + s.BackendPackage
	}

	s.PackageName = gstr.TrimLeftStr(s.BackendPackage, goModuleName+"/")
	tableName := s.Name
	if g.IsEmpty(s.BusinessName) {
		s.BusinessName = tableName
	}
	s.PackageNameProto = gstr.ToLower(gstr.Replace(s.BackendPackage, "/", "."))
	s.ClassName = gstr.CaseCamel(s.BusinessName)
	s.StructName = gstr.CaseCamelLower(s.BusinessName)
	s.GoFileName = gstr.CaseSnake(s.BusinessName)
	s.RouteChildPath = gstr.CaseKebab(s.BusinessName)
	s.FrontendFileName = gstr.CaseKebab(s.BusinessName)
	s.FrontendPath = gstr.CaseKebab(s.FrontendModule)
}

func (s *TableDef) ProcessColumns(ctx context.Context, yamlInputPath string, goModuleName string, cache map[string]*TableDef) (err error) {
	for _, column := range s.Columns {
		if g.IsEmpty(column.Name) {
			return gerror.Newf("表%s中的字段没有给定name", s.Name)
		}
		if err = column.SetColumnValues(); err != nil {
			return err
		}
		if column.IsPk {
			s.PkColumns[column.Name] = column
		}
		if column.GoType == "Time" {
			s.HasTimeColumnInMain = true
			s.HasTimeColumn = true
		}
		if column.HtmlType == "images" || column.HtmlType == "file" || column.HtmlType == "files" {
			s.HasUpFileColumn = true
		}
		if column.HtmlType == "checkbox" {
			s.HasCheckboxColumn = true
		}
	}
	for _, column := range s.VirtualColumns {
		if err = column.SetColumnValues(); err != nil {
			return err
		}
	}

	for _, addColumn := range s.AddColumns {
		columnName := addColumn.Name
		baseColumn, found := s.ColumnMap[columnName]
		if !found {
			return gerror.Newf("新增字段 %s 不存在于表 %s 的 columns 定义中", s.Name, columnName)
		}
		s.SetAddColumnValues(addColumn, baseColumn)
	}
	isPkInEdit := false
	for _, editColumn := range s.EditColumns {
		columnName := editColumn.Name
		baseColumn, found := s.ColumnMap[columnName]
		if !found {
			return gerror.Newf("编辑字段 %s 不存在于表 %s 的 columns 定义中", s.Name, columnName)
		}
		if baseColumn.IsPk {
			isPkInEdit = true
		}
		s.SetEditColumnValues(editColumn, baseColumn)
	}
	s.IsPkInEdit = isPkInEdit
	for _, listColumn := range s.ListColumns {
		columnName := listColumn.Name
		baseColumn, found := s.ColumnMap[columnName]
		if !found {
			baseColumn, found = s.VirtualColumnMap[columnName]
			if !found {
				return gerror.Newf("列表字段 %s 不存在于表 %s 的 columns 和 virtualColumns 定义中", s.Name, columnName)
			}
		}
		s.SetListColumnValues(listColumn, baseColumn)
	}
	for _, detailColumn := range s.DetailColumns {
		columnName := detailColumn.Name
		baseColumn, found := s.ColumnMap[columnName]
		if !found {
			baseColumn, found = s.VirtualColumnMap[columnName]
			if !found {
				return gerror.Newf("详情字段 %s 不存在于表 %s 的 columns 和 virtualColumns 定义中", s.Name, columnName)
			}
		}
		s.SetDetailColumnValues(detailColumn, baseColumn)
	}
	for _, queryColumn := range s.QueryColumns {
		columnName := queryColumn.Name
		baseColumn, found := s.ColumnMap[columnName]
		if !found {
			baseColumn, found = s.VirtualColumnMap[columnName]
			if !found {
				return gerror.Newf("查询字段 %s 不存在于表 %s 的 columns 和 virtualColumns 定义中", s.Name, columnName)
			}
		}
		hasConversion := s.SetQueryColumnValues(queryColumn, baseColumn)
		s.HasConversion = s.HasConversion || hasConversion
		if baseColumn.IsVirtual {
			s.HasVirtualQueries = true
			foreignTableName := baseColumn.ForeignTableName
			foreignTable, err1 := LoadTableDefYaml(ctx, foreignTableName, yamlInputPath, goModuleName, cache)
			if err1 != nil {
				return err1
			}
			s.VirtualQueryRelated[foreignTableName] = foreignTable
		}
	}
	return
}

func (c *ColumnDef) SetColumnValues() error {
	if g.IsEmpty(c.SqlType) {
		return gerror.Newf("字段%s必须给定sqlType", c.Name)
	}
	dataType, isUnsigned := GetDataType(c.SqlType)
	columnName := c.Name
	//设置字段名
	if g.IsEmpty(c.GoField) {
		c.GoField = gstr.CaseCamel(columnName)
	}
	if g.IsEmpty(c.HtmlField) {
		c.HtmlField = gstr.CaseCamelLower(columnName)
	}

	if g.IsEmpty(c.GoType) {
		if IsStringObject(dataType) {
			c.GoType = "string"
		} else if IsTimeObject(dataType) || IsDateObject(dataType) {
			c.GoType = "Time"
		} else if IsNumberObject(dataType) {
			switch dataType {
			case "float", "double", "decimal", "numeric":
				c.GoType = "float64"
			case "int", "integer", "tinyint", "smallint", "mediumint":
				if isUnsigned {
					c.GoType = "uint32"
				} else {
					c.GoType = "int32"
				}
			case "bigint":
				if isUnsigned {
					c.GoType = "uint64"
				} else {
					c.GoType = "int64"
				}
			case "bit":
				c.GoType = "bool"
			}
		} else {
			c.GoType = "string"
		}
	}
	c.ConvertFunc = gstr.CaseCamel(c.GoType)

	if c.GoType == "int" {
		c.GoType = "int32"
	} else if c.GoType == "uint" {
		c.GoType = "uint32"
	}

	switch c.GoType {
	case "Time":
		c.ProtoType = "string"
	case "float64":
		c.ProtoType = "double"
	case "int":
		c.ProtoType = "int32"
	case "uint":
		c.ProtoType = "uint32"
	default:
		c.ProtoType = c.GoType
	}

	if g.IsEmpty(c.HtmlType) {
		if IsStringObject(dataType) {
			columnLength := GetColumnLength(c.SqlType)
			if columnLength >= 500 {
				c.HtmlType = "textarea"
			} else {
				c.HtmlType = "input"
			}
		} else if IsDateObject(dataType) {
			c.HtmlType = "date"
		} else if IsTimeObject(dataType) {
			c.HtmlType = "datetime"
		} else if IsNumberObject(dataType) {
			c.HtmlType = "input"
		} else if dataType == "bit" {
			c.HtmlType = "select"
		} else {
			c.HtmlType = "input"
		}
	}
	return nil
}

func (s *TableDef) SetAddColumnValues(addColumn *AddColumnDef, baseColumn *ColumnDef) {
	addColumn.Base = baseColumn
	addColumn.Comment = baseColumn.Comment
	addColumn.GoType = baseColumn.GoType
	addColumn.GoField = baseColumn.GoField
	addColumn.HtmlField = baseColumn.HtmlField
	if g.IsEmpty(addColumn.HtmlType) {
		addColumn.HtmlType = baseColumn.HtmlType
	}
}

func (s *TableDef) SetEditColumnValues(editColumn *EditColumnDef, baseColumn *ColumnDef) {
	editColumn.Base = baseColumn
	editColumn.Comment = baseColumn.Comment
	editColumn.GoType = baseColumn.GoType
	editColumn.GoField = baseColumn.GoField
	editColumn.HtmlField = baseColumn.HtmlField
	if g.IsEmpty(editColumn.HtmlType) {
		editColumn.HtmlType = baseColumn.HtmlType
	}
	if baseColumn.IsPk {
		editColumn.IsDisabled = true
	}
}

func (s *TableDef) SetListColumnValues(listColumn *ListColumnDef, baseColumn *ColumnDef) {
	listColumn.Base = baseColumn
	listColumn.Comment = baseColumn.Comment
	listColumn.GoType = baseColumn.GoType
	listColumn.GoField = baseColumn.GoField
	listColumn.HtmlField = baseColumn.HtmlField
	if g.IsEmpty(listColumn.HtmlType) {
		listColumn.HtmlType = baseColumn.HtmlType
	}
}

func (s *TableDef) SetQueryColumnValues(queryColumn *QueryColumnDef, baseColumn *ColumnDef) (hasConversion bool) {
	columnName := baseColumn.Name
	queryColumn.Base = baseColumn
	queryColumn.Comment = baseColumn.Comment
	queryColumn.GoType = baseColumn.GoType
	queryColumn.GoField = baseColumn.GoField
	queryColumn.HtmlField = baseColumn.HtmlField
	if g.IsEmpty(queryColumn.HtmlType) {
		queryColumn.HtmlType = baseColumn.HtmlType
	}
	if g.IsEmpty(queryColumn.QueryType) {
		queryColumn.QueryType = "EQ"
	}
	// validation 规则 和 conversion 方法
	integerValidationRule := "integer"
	floatValidationRule := "float"
	dateValidationRule := "date"
	datetimeValidationRule := "date-format:Y-m-d H:i:s"
	if queryColumn.QueryType == "BETWEEN" {
		integerValidationRule += "-array"
		floatValidationRule += "-array"
		dateValidationRule += "-array"
		datetimeValidationRule += "-array"
	}
	hasConversion = false
	switch baseColumn.GoType {
	case "int":
		queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@" + integerValidationRule + "#" + baseColumn.Comment + "需为整数"
		queryColumn.FieldConversion = "gconv.Int"
		hasConversion = true
		break
	case "int64":
		queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@" + integerValidationRule + "#" + baseColumn.Comment + "需为整数"
		queryColumn.FieldConversion = "gconv.Int64"
		hasConversion = true
		break
	case "uint":
		queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@" + integerValidationRule + "#" + baseColumn.Comment + "需为整数"
		queryColumn.FieldConversion = "gconv.Uint"
		hasConversion = true
		break
	case "uint64":
		queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@" + integerValidationRule + "#" + baseColumn.Comment + "需为整数"
		queryColumn.FieldConversion = "gconv.Uint64"
		hasConversion = true
		break
	case "float":
		queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@" + floatValidationRule + "#" + baseColumn.Comment + "需为浮点数"
		queryColumn.FieldConversion = "gconv.Float"
		hasConversion = true
		break
	case "float64":
		queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@" + floatValidationRule + "#" + baseColumn.Comment + "需为浮点数"
		queryColumn.FieldConversion = "gconv.Float64"
		hasConversion = true
		break
	case "bool":
		queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@boolean#" + baseColumn.Comment + "需为true/false"
		queryColumn.FieldConversion = "gconv.Bool"
		hasConversion = true
		break
	case "Time":
		if baseColumn.HtmlType == "date" {
			queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@" + dateValidationRule + "#" + baseColumn.Comment + "需为YYYY-MM-DD格式"
			queryColumn.FieldConversion = "gconv.Time"
		} else {
			queryColumn.FieldValidation = gstr.CaseCamelLower(columnName) + "@" + datetimeValidationRule + "#" + baseColumn.Comment + "需为YYYY-MM-DD hh:mm:ss格式"
			queryColumn.FieldConversion = "gconv.Time"
		}
		hasConversion = true
		break
	}
	return
}

func (s *TableDef) SetDetailColumnValues(detailColumn *DetailColumnDef, baseColumn *ColumnDef) {
	detailColumn.Base = baseColumn
	detailColumn.Comment = baseColumn.Comment
	detailColumn.GoType = baseColumn.GoType
	detailColumn.GoField = baseColumn.GoField
	detailColumn.HtmlField = baseColumn.HtmlField
	if g.IsEmpty(detailColumn.HtmlType) {
		detailColumn.HtmlType = baseColumn.HtmlType
	}
}

func (s *TableDef) ProcessCascadeColumn(column *ColumnDef) (err error) {
	if !column.IsCascade {
		return
	}
	parentColumnName := column.ParentColumnName
	if g.IsEmpty(parentColumnName) {
		err = gerror.New("级联查询字段\"" + column.Name + "\"并未设置parentColumnName")
		return
	}
	parent, found := s.ColumnMap[parentColumnName]
	if !found {
		parent, found = s.VirtualColumnMap[parentColumnName]
		if !found {
			err = gerror.New("级联查询字段\"" + column.Name + "\"的parentColumnName\"" + parentColumnName + "\"不存在于表中")
			return
		}
	}
	column.ParentColumnName = parentColumnName
	column.CascadeParent = parent
	parent.IsCascadeParent = true
	parent.CascadeChildrenColumns = gset.NewStrSet()
	return
}

func (s *TableDef) AddChildren(column *ColumnDef) (err error) {
	if !column.IsCascade {
		return
	}
	child := column
	for {
		parentColumnName := child.ParentColumnName
		if g.IsEmpty(parentColumnName) {
			break
		}
		parent, found := s.ColumnMap[parentColumnName]
		if !found {
			parent, found = s.VirtualColumnMap[parentColumnName]
			if !found {
				err = gerror.New("级联查询字段\"" + column.Name + "\"的parentColumnName\"" + parentColumnName + "\"不存在于表中")
				return
			}
		}
		parent.CascadeChildrenColumns.Add(column.Name)
		child = parent
	}
	return
}

func (s *TableDef) ProcessCascades() error {
	for _, column := range s.Columns {
		err := s.ProcessCascadeColumn(column)
		if err != nil {
			return err
		}
	}
	for _, column := range s.VirtualColumns {
		err := s.ProcessCascadeColumn(column)
		if err != nil {
			return err
		}
	}
	for _, column := range s.Columns {
		err := s.AddChildren(column)
		if err != nil {
			return err
		}
	}
	for _, column := range s.VirtualColumns {
		err := s.AddChildren(column)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TableDef) ProcessRelatedAndForeign(ctx context.Context, yamlInputPath string, goModuleName string, cache map[string]*TableDef) error {
	for _, column := range s.Columns {
		err := s.ProcessColumnRelatedAndForeign(ctx, column, yamlInputPath, goModuleName, cache)
		if err != nil {
			return err
		}
	}
	for _, column := range s.VirtualColumns {
		err := s.ProcessColumnRelatedAndForeign(ctx, column, yamlInputPath, goModuleName, cache)
		if err != nil {
			return err
		}
	}

	if s.FkColumnNameSet != nil {
		for _, fkColumnName := range s.FkColumnNameSet.Slice() {
			if !s.IsInList(fkColumnName) {
				s.FkColumnsNotInList = append(s.FkColumnsNotInList, s.ColumnMap[fkColumnName])
			}
		}
	}

	if s.RelatedTableMap != nil {
		s.RelatedTables = s.RelatedTableMap.Values()
		for _, relatedTable := range s.RelatedTables {
			s.HasTimeColumn = s.HasTimeColumn || relatedTable.(*TableDef).HasTimeColumn
			s.HasUpFileColumn = s.HasUpFileColumn || relatedTable.(*TableDef).HasUpFileColumn
			if relatedTable.(*TableDef).RelatedTableMap == nil {
				continue
			}
			relatedTable.(*TableDef).RelatedTables = relatedTable.(*TableDef).RelatedTableMap.Values()
			for _, innerRelatedTable := range relatedTable.(*TableDef).RelatedTables {
				s.HasTimeColumn = s.HasTimeColumn || innerRelatedTable.(*TableDef).HasTimeColumn
				s.HasUpFileColumn = s.HasUpFileColumn || innerRelatedTable.(*TableDef).HasUpFileColumn
			}
		}
	}

	if s.AllRelatedTableMap != nil {
		s.AllRelatedTables = s.AllRelatedTableMap.Values()
	}

	return nil
}

func (s *TableDef) ProcessColumnRelatedAndForeign(ctx context.Context, column *ColumnDef, yamlInputPath string, goModuleName string, cache map[string]*TableDef) error {
	if g.IsEmpty(column.RelatedTableName) && g.IsEmpty(column.ForeignTableName) {
		return nil
	}
	if !g.IsEmpty(column.ForeignTableName) { // 存在外表（虚拟字段所在的表）
		foreignTable, err := s.AddRelatedInfo(ctx, column.ForeignTableName, column.ForeignValueColumnName, column.ForeignKeyColumnName, yamlInputPath, goModuleName, cache)
		if err != nil {
			return err
		}
		// RltdBaseTableNameForeignTableName
		foreignTable.ClassNameWhenRelated = RelatedTablePrefix + s.ClassName + foreignTable.ClassName
		foreignTable.JsonNameWhenRelated = RelatedTableJsonPrefix + s.ClassName + foreignTable.ClassName
		foreignTable.CombinedClassName = foreignTable.ClassName
		column.ForeignTableClass = foreignTable.ClassName

		if !g.IsEmpty(column.RelatedTableName) { // 主表->外表->外表的关联表 三级嵌套
			innerRelatedTable, err1 := foreignTable.AddRelatedInfo(ctx, column.RelatedTableName, column.RelatedValueColumnName, column.ForeignValueColumnName, yamlInputPath, goModuleName, cache)
			if err1 != nil {
				return err
			}
			innerRelatedTable.ClassNameWhenRelated = RelatedTablePrefix + s.ClassName + foreignTable.ClassName + innerRelatedTable.ClassName
			innerRelatedTable.JsonNameWhenRelated = RelatedTableJsonPrefix + s.ClassName + foreignTable.ClassName + innerRelatedTable.ClassName
			innerRelatedTable.CombinedClassName = foreignTable.ClassName + innerRelatedTable.ClassName
			column.RelatedKeyColumn = innerRelatedTable.PkColumns
			column.CombinedTableClass = innerRelatedTable.ClassNameWhenRelated
			column.CombinedHtmlTableClass = foreignTable.ClassName + innerRelatedTable.ClassName
			column.CombinedHtmlField = foreignTable.JsonNameWhenRelated + "." +
				innerRelatedTable.JsonNameWhenRelated + "." + gstr.CaseCamelLower(column.RelatedValueColumnName)

			if s.AllRelatedTableMap == nil {
				s.AllRelatedTableMap = gmap.NewListMap()
			}
			s.AllRelatedTableMap.Set(innerRelatedTable.CombinedClassName, innerRelatedTable)
		} else { // 主表->外表 两级嵌套
			column.RelatedKeyColumn = foreignTable.PkColumns
			column.CombinedTableClass = foreignTable.ClassNameWhenRelated
			column.CombinedHtmlTableClass = foreignTable.ClassName
			column.CombinedHtmlField = foreignTable.JsonNameWhenRelated + "." + gstr.CaseCamelLower(column.ForeignValueColumnName)

			if s.AllRelatedTableMap == nil {
				s.AllRelatedTableMap = gmap.NewListMap()
			}
			s.AllRelatedTableMap.Set(foreignTable.CombinedClassName, foreignTable)
		}

		if s.FkColumnNameSet == nil {
			s.FkColumnNameSet = gset.NewStrSet()
		}
		s.FkColumnNameSet.Add(column.ForeignKeyColumnName)

	} else if !g.IsEmpty(column.RelatedTableName) { // 不存在外表，只有主表->主表的关联表 两级嵌套
		relatedTable, err := s.AddRelatedInfo(ctx, column.RelatedTableName, column.RelatedValueColumnName, column.Name, yamlInputPath, goModuleName, cache)
		if err != nil {
			return err
		}
		relatedTable.ClassNameWhenRelated = RelatedTablePrefix + s.ClassName + relatedTable.ClassName
		relatedTable.JsonNameWhenRelated = RelatedTableJsonPrefix + s.ClassName + relatedTable.ClassName
		relatedTable.CombinedClassName = relatedTable.ClassName
		column.RelatedKeyColumn = relatedTable.PkColumns
		column.CombinedTableClass = relatedTable.ClassNameWhenRelated
		column.CombinedHtmlTableClass = relatedTable.ClassName
		column.CombinedHtmlField = relatedTable.JsonNameWhenRelated + "." + gstr.CaseCamelLower(column.RelatedValueColumnName)

		if s.AllRelatedTableMap == nil {
			s.AllRelatedTableMap = gmap.NewListMap()
		}
		s.AllRelatedTableMap.Set(relatedTable.CombinedClassName, relatedTable)
	}
	return nil
}

func (s *TableDef) AddRelatedInfo(ctx context.Context, relatedTableName, relatedValueColumnName, originalColumnName string, yamlInputPath string, goModuleName string, cache map[string]*TableDef) (*TableDef, error) {
	if s.RelatedTableMap == nil {
		s.RelatedTableMap = &gmap.ListMap{}
	}
	relatedTable := s.RelatedTableMap.GetOrSetFunc(relatedTableName, func() interface{} {
		t, err := LoadTableDefYaml(ctx, relatedTableName, yamlInputPath, goModuleName, cache)
		if err != nil {
			return err
		}
		return t
	}).(*TableDef)
	err := relatedTable.AddWithInfo(ctx, relatedValueColumnName, originalColumnName)
	if err != nil {
		return nil, err
	}
	return relatedTable, nil
}

func (s *TableDef) AddWithInfo(ctx context.Context, destValueColumn, originalColumn string) error {
	relatedValueColumn, foundValue := s.ColumnMap[destValueColumn]
	if !foundValue {
		return gerror.Newf("无法找到关联表的列 %s", destValueColumn)
	}

	var (
		pkColumnName string
		pkColumn     *ColumnDef
	)
	if !g.IsEmpty(s.PkColumns) {
		if len(s.PkColumns) > 1 {
			g.Log().Warningf(ctx, "当前表%s定义了多个主键列，无法在表关联时被自动引用查询", s.Name)
		} else {
			for key, val := range s.PkColumns {
				pkColumnName = key
				pkColumn = val
				break
			}
			s.OrmWithMapping = "orm:\"with:" + pkColumnName + "=" + originalColumn + "\""
			s.RefColumns.GetOrSet(pkColumnName, pkColumn)
		}
	} else {
		g.Log().Warningf(ctx, "当前表%s没有定义主键列，无法在表关联时被自动引用查询", s.Name)
	}
	s.RefColumns.GetOrSet(relatedValueColumn.Name, &relatedValueColumn)
	return nil
}

func (s *TableDef) IsInList(columnName string) bool {
	for _, c := range s.ListColumns {
		if c.Name == columnName {
			return true
		}
	}
	return false
}
