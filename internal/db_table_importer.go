package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"strings"
)

var DbTableImporter = new(dbTableImporter)

type dbTableImporter struct{}

func (s *dbTableImporter) GenDbTableDefs(ctx context.Context, importOptions *ImportOptions) error {
	tableNames := importOptions.TableNames
	tablePrefixesOnly := importOptions.TablePrefixesOnly
	tables, err := s.getDbTablesByNames(ctx, tableNames, tablePrefixesOnly)
	if err != nil {
		g.Log().Error(ctx, err)
		return err
	}
	for _, table := range tables {
		err = s.fillTableDef(ctx, table, importOptions.GoModuleName)
		if !g.IsEmpty(importOptions.BackendPackage) {
			table.BackendPackage = importOptions.BackendPackage
		}
		if !g.IsEmpty(importOptions.FrontendModule) {
			table.FrontendModule = importOptions.FrontendModule
		}
		if !g.IsEmpty(importOptions.SeparatePackage) {
			table.SeparatePackage = importOptions.SeparatePackage
		}
		if !g.IsEmpty(importOptions.TemplateCategory) {
			table.TemplateCategory = importOptions.TemplateCategory
		}
		if !g.IsEmpty(importOptions.Author) {
			table.FunctionAuthor = importOptions.Author
		}
		if !g.IsEmpty(importOptions.Overwrite) {
			table.Overwrite = importOptions.Overwrite
		}
		if !g.IsEmpty(importOptions.ShowDetail) {
			table.ShowDetail = importOptions.ShowDetail
		}
		if !g.IsEmpty(importOptions.IsRpc) {
			table.IsRpc = importOptions.IsRpc
		}
		if !g.IsEmpty(importOptions.Prefix) {
			table.BusinessName = strings.Replace(table.BusinessName, importOptions.Prefix, "", 1)
		}
		if err != nil {
			g.Log().Error(ctx, err)
			return err
		}
		err = saveTableDef(ctx, table, importOptions.YamlOutputPath)
		if err != nil {
			g.Log().Error(ctx, err)
			return err
		}
	}
	return nil
}

func (s *dbTableImporter) getDbTablesByNames(ctx context.Context, tableNames []string, prefixes []string) ([]*tableDef, error) {
	if getDbDriver() != "mysql" {
		return nil, gerror.New("代码生成只支持mysql数据库")
	}
	db := g.DB(gdb.DefaultGroupName)
	sql := "select TABLE_NAME as name, TABLE_COMMENT as comment" +
		"     from information_schema.tables" +
		"    where table_name NOT LIKE 'qrtz_%'" +
		"      and table_name NOT LIKE 'gen_%' " +
		"      and table_schema = (select database()) "
	if len(tableNames) > 0 {
		in := gstr.TrimRight(gstr.Repeat("?,", len(tableNames)), ",")
		sql += " and " + gdb.FormatSqlWithArgs("table_name in ("+in+")", gconv.SliceAny(tableNames))
	}
	if len(prefixes) > 0 {
		sql += " and ("
		for i, prefix := range prefixes {
			if i > 0 {
				sql += " or "
			}
			sql += " table_name like ('" + prefix + "%')"
		}
		sql += ")"
	}
	var result []*tableDef
	err := db.GetScan(ctx, &result, sql)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *dbTableImporter) fillTableDef(ctx context.Context, table *tableDef, goModuleName string) error {
	tableName := table.Name
	// 保存列信息
	columns, err := s.selectDbTableColumnsByName(ctx, tableName)
	if err != nil || len(columns) <= 0 {
		return gerror.New("获取列数据失败")
	}
	s.setTableDefaults(table, goModuleName, len(columns))
	for _, column := range columns {
		//s.setColumnDefaults(column)
		columnName := column.Name
		if column.IsPk {
			table.PkColumn = column
			table.SortColumn = columnName
		}
		if columnName == "created_at" {
			table.CreatedAtColumn = column
		}
		if columnName == "created_by" {
			table.HasCreatedBy = true
			table.CreatedByColumn = column
		}
		if columnName == "updated_by" {
			table.HasUpdatedBy = true
		}
		listColumnDefault := s.getListColumnDefault(column)
		addColumnDefault := s.getAddColumnDefault(column)
		editColumnDefault := s.getEditColumnDefault(column)
		queryColumnDefault := s.getQueryColumnDefault(column)
		detailColumnDefault := s.getDetailColumnDefault(column)

		if !column.IsPk || !column.IsIncrement {
			table.AddColumns = append(table.AddColumns, addColumnDefault)
			if column.IsPk {
				editColumnDefault.IsDisabled = true
			}
			table.EditColumns = append(table.EditColumns, editColumnDefault)
		}
		table.ListColumns = append(table.ListColumns, listColumnDefault)
		table.QueryColumns = append(table.QueryColumns, queryColumnDefault)
		table.DetailColumns = append(table.DetailColumns, detailColumnDefault)
		table.Columns = append(table.Columns, column)
	}
	return nil
}

// selectDbTableColumnsByName 根据表名称查询列信息
func (s *dbTableImporter) selectDbTableColumnsByName(ctx context.Context, tableName string) ([]*columnDef, error) {
	db := g.DB(gdb.DefaultGroupName)
	var res []*columnDef
	sql := " select column_name as name," +
		"           (case when (is_nullable = 'YES' || is_nullable = 'NO' && column_default is not null) then '0' else '1' end) as is_required," +
		"           (case when column_key = 'PRI' then '1' else '0' end) as is_pk," +
		"           ordinal_position as sort," +
		"           column_comment as comment," +
		"           (case when extra = 'auto_increment' then '1' else '0' end) as is_increment," +
		"           column_type as sql_type" +
		"      from information_schema.columns" +
		"     where table_schema = (select database()) "
	sql += " and " + gdb.FormatSqlWithArgs(" table_name=? ", []interface{}{tableName}) + " order by ordinal_position ASC "
	err := db.GetScan(ctx, &res, sql)
	if err != nil {
		return nil, gerror.New("查询列信息失败")
	}
	return res, nil
}

// InitTable 初始化表信息
func (s *dbTableImporter) setTableDefaults(table *tableDef, goModuleName string, columnCount int) {
	table.setVariableNames(goModuleName)
	table.FunctionName = strings.ReplaceAll(table.Comment, "表", "")
	table.TemplateCategory = "crud"
	table.SortType = "asc"
	table.CreateTime = gtime.Now()
	table.UpdateTime = table.CreateTime
	table.ColumnMap = make(map[string]*columnDef, columnCount)
	table.Columns = []*columnDef{}
	table.ListColumns = []*listColumnDef{}
	table.AddColumns = []*addColumnDef{}
	table.EditColumns = []*editColumnDef{}
	table.QueryColumns = []*queryColumnDef{}
	table.DetailColumns = []*detailColumnDef{}
}

func (s *dbTableImporter) setColumnDefaults(column *columnDef) {
	dataType := column.SqlType
	columnName := column.Name
	//设置字段名
	if g.IsEmpty(column.GoField) {
		column.GoField = gstr.CaseCamel(columnName)
	}
	if g.IsEmpty(column.HtmlField) {
		column.HtmlField = gstr.CaseCamelLower(columnName)
	}

	if g.IsEmpty(column.GoType) {
		if IsStringObject(dataType) {
			column.GoType = "string"
		} else if IsTimeObject(dataType) || IsDateObject(dataType) {
			column.GoType = "Time"
		} else if IsNumberObject(dataType) {
			t, _ := gregex.ReplaceString(`\(.+\)`, "", column.SqlType)
			t = gstr.Split(gstr.Trim(t), " ")[0]
			t = gstr.ToLower(t)
			// 如果是浮点型
			switch t {
			case "float", "double", "decimal":
				column.GoType = "float64"
			case "bit", "int", "tinyint", "small_int", "smallint", "medium_int", "mediumint":
				if gstr.ContainsI(column.SqlType, "unsigned") {
					column.GoType = "uint"
				} else {
					column.GoType = "int"
				}
			case "big_int", "bigint":
				if gstr.ContainsI(column.SqlType, "unsigned") {
					column.GoType = "uint64"
				} else {
					column.GoType = "int64"
				}
			}
		} else if dataType == "bit" {
			column.GoType = "bool"
		}
	}

	if g.IsEmpty(column.HtmlType) {
		if IsStringObject(dataType) {
			columnLength := GetColumnLength(column.SqlType)
			if columnLength >= 500 {
				column.HtmlType = "textarea"
			} else {
				column.HtmlType = "input"
			}
		} else if IsDateObject(dataType) {
			column.HtmlType = "date"
		} else if IsTimeObject(dataType) {
			column.HtmlType = "datetime"
		} else if IsNumberObject(dataType) {
			column.HtmlType = "input"
		} else if dataType == "bit" {
			column.HtmlType = "select"
		}
	}
	return
}

func (s *dbTableImporter) getAddColumnDefault(base *columnDef) *addColumnDef {
	return &addColumnDef{
		Name: base.Name,
		Sort: base.Sort,
	}
}

func (s *dbTableImporter) getEditColumnDefault(base *columnDef) *editColumnDef {
	return &editColumnDef{
		Name: base.Name,
		Sort: base.Sort,
	}
}

func (s *dbTableImporter) getListColumnDefault(base *columnDef) *listColumnDef {
	return &listColumnDef{
		Name:              base.Name,
		Sort:              base.Sort,
		IsOverflowTooltip: true,
		MinWidth:          100,
	}
}

func (s *dbTableImporter) getQueryColumnDefault(base *columnDef) (queryColumn *queryColumnDef) {
	return &queryColumnDef{
		Name: base.Name,
		Sort: base.Sort,
	}
}

func (s *dbTableImporter) getDetailColumnDefault(base *columnDef) *detailColumnDef {
	return &detailColumnDef{
		Name:    base.Name,
		Sort:    base.Sort,
		ColSpan: 12,
	}
}
