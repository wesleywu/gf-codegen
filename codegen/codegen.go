package main

import (
	"context"
	"github.com/WesleyWu/gf-codegen/internal"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func CodeGenFunc(ctx context.Context, parser *gcmd.Parser) error {
	tablesStr := parser.GetOpt("tables").String()
	tablePrefixOnlyStr := parser.GetOpt("tablePrefixOnly").String()
	yamlInputPath := parser.GetOpt("yamlInputPath", "manifest/config/codegen_conf").String()
	serviceOnly := parser.GetOpt("serviceOnly").Bool()
	frontendType := parser.GetOpt("frontendType").String()
	frontendPath := parser.GetOpt("frontendPath").String()

	tableNamesFilter := gset.NewStrSetFrom(internal.SplitComma(tablesStr))
	tablePrefixesOnly := internal.SplitComma(tablePrefixOnlyStr)
	goModuleName, err := internal.GetGoModuleName()
	if err != nil {
		return err
	}

	genOption := &internal.GenOptions{
		YamlInputPath: yamlInputPath,
		GoModuleName:  goModuleName,
		ServiceOnly:   serviceOnly,
		FrontendType:  frontendType,
		FrontendPath:  frontendPath,
	}

	curDir, err := os.Getwd()
	if err != nil {
		return gerror.Wrap(err, "获取本地路径失败")
	}
	yamlPath := path.Join(curDir, yamlInputPath)
	fileList, err := ioutil.ReadDir(yamlPath)
	if err != nil {
		return gerror.Wrap(err, "读取目录出错")
	}
	var tableNames []string
	for _, file := range fileList {
		if file.IsDir() {
			continue
		}
		pos := gstr.PosR(file.Name(), ".yaml")
		if pos == -1 {
			continue
		}
		tableName := file.Name()[0:pos]
		if tableNamesFilter.Size() > 0 && !tableNamesFilter.Contains(tableName) {
			continue
		}
		if len(tablePrefixesOnly) > 0 {
			matchPrefix := false
			for _, onePrefix := range tablePrefixesOnly {
				if gstr.Pos(tableName, onePrefix) > 0 {
					matchPrefix = true
					break
				}
			}
			if !matchPrefix {
				continue
			}
		}
		tableNames = append(tableNames, tableName)
	}

	for _, tableName := range tableNames {
		g.Log().Infof(ctx, "generating code for table %s in go module %s", tableName, genOption.GoModuleName)
		err = internal.GenCodeByTableDefYaml(ctx, tableName, genOption)
		if err != nil {
			return err
		}
		g.Log().Info(ctx, "done")
	}
	if err != nil {
		return err
	}
	if !serviceOnly {
		g.Log().Info(ctx, "importing http helpers from github.com/WesleyWu/gf-httputils")
		cmd := exec.Command("go", "get", "github.com/WesleyWu/gf-httputils")
		bytes, err := cmd.Output()
		if err != nil {
			return err
		}
		if len(bytes) > 0 {
			g.Log().Info(ctx, string(bytes))
		} else {
			g.Log().Info(ctx, "done")
		}
		g.Log().Info(ctx, "executing go mod tidy")
		cmd = exec.Command("go", "mod", "tidy")
		bytes, err = cmd.Output()
		if err != nil {
			return err
		}
		if len(bytes) > 0 {
			g.Log().Info(ctx, string(bytes))
		} else {
			g.Log().Info(ctx, "done")
		}
	}
	return nil
}

func main() {
	command := gcmd.Command{
		Name: "Code gen",
		Func: CodeGenFunc,
	}
	command.Run(gctx.New())
}
