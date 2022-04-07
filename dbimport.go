package main

import (
	"context"
	"github.com/WesleyWu/gf-codegen/internal"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
)

func ImportFunc(ctx context.Context, parser *gcmd.Parser) error {
	dblink := parser.GetOpt("dblink").String()
	tablesStr := parser.GetOpt("tables").String()
	tablePrefixOnlyStr := parser.GetOpt("tablePrefixOnly").String()
	removeTablePrefixStr := parser.GetOpt("removeTablePrefix").String()
	backendPackage := parser.GetOpt("backendPackage").String()
	frontendModule := parser.GetOpt("frontendModule").String()
	yamlOutputPath := parser.GetOpt("yamlOutputPath", "manifest/config/codegen_conf").String()
	separatePackage := parser.GetOpt("separatePackage", false).Bool()
	author := parser.GetOpt("author", "Awesome Developer").String()
	overwrite := parser.GetOpt("overwrite", true).Bool()
	showDetail := parser.GetOpt("showDetail", true).Bool()
	isRpc := parser.GetOpt("isRpc", false).Bool()

	err := internal.ParseDblink(dblink)
	if err != nil {
		return err
	}

	tables := internal.SplitComma(tablesStr)
	removeTablePrefixes := internal.SplitComma(removeTablePrefixStr)
	tablePrefixesOnly := internal.SplitComma(tablePrefixOnlyStr)
	goModuleName, err := internal.GetGoModuleName()
	if err != nil {
		return err
	}

	backendPackage = gstr.TrimLeftStr(backendPackage, "/")
	backendPackage = gstr.TrimRightStr(backendPackage, "/")

	if gstr.Pos(backendPackage, goModuleName) != 0 {
		backendPackage = goModuleName + "/" + backendPackage
	}

	importOptions := &internal.ImportOptions{
		BackendPackage:      backendPackage,
		FrontendModule:      frontendModule,
		GoModuleName:        goModuleName,
		TableNames:          tables,
		RemoveTablePrefixes: removeTablePrefixes,
		TablePrefixesOnly:   tablePrefixesOnly,
		SeparatePackage:     separatePackage,
		Author:              author,
		Overwrite:           overwrite,
		ShowDetail:          showDetail,
		IsRpc:               isRpc,
		YamlOutputPath:      yamlOutputPath,
	}

	err = internal.DbTableImporter.GenDbTableDefs(gctx.New(), importOptions)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	command := gcmd.Command{
		Name: "Database schema importer",
		Func: ImportFunc,
	}
	command.Run(gctx.New())
}
