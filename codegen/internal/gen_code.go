package internal

import (
	"bufio"
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/WesleyWu/gf-codegen/common"
	"github.com/WesleyWu/gf-codegen/common/protobuf"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"io"
	"os"
	"strings"
)

//go:embed template/go/controller.template
var controllerTemplate string

//go:embed template/go/dao.template
var daoTemplate string

//go:embed template/go/dao_internal.template
var daoInternalTemplate string

//go:embed template/go/entity.template
var entityTemplate string

//go:embed template/go/model.template
var modelTemplate string

//go:embed template/go/provider.template
var providerTemplate string

//go:embed template/go/router.template
var routerTemplate string

//go:embed template/go/service.template
var serviceTemplate string

//go:embed template/js/api.template
var jsapiTemplate string

//go:embed template/protobuf/protobuf.template
var protobufTemplate string

//go:embed template/sql/sql.template
var sqlTemplate string

//go:embed template/vue/list-vue.template
var listVueTemplate string

//go:embed template/vue/tree-vue.template
var treeVueTemplate string

func GenCodeByTableDefYaml(ctx context.Context, tableName string, genOptions *common.GenOptions) error {
	var cache = map[string]*common.TableDef{}
	table, err := common.LoadTableDefYaml(ctx, tableName, genOptions.YamlInputPath, genOptions.GoModuleName, cache)
	if err != nil {
		return err
	}
	if table.IsRpc {
		// make sure protoc can work properly
		protocVersionOk, err1 := protobuf.IsProtocVersionOK()
		if err1 != nil {
			return err
		}
		if !protocVersionOk {
			return gerror.New("请安装3.0以上版本的 protoc")
		}
		protocTripleVersionOk, err2 := protobuf.IsProtocTripleVersionOK()
		if err2 != nil {
			return err
		}
		if !protocTripleVersionOk {
			return gerror.New("请安装1.0以上版本的 protoc-gen-go-triple")
		}
		if g.IsEmpty(table.RpcPort) {
			return gerror.New("必须指定rpc服务侦听端口 RpcPort，建议20000以上，各服务的端口号不能重复")
		}
	}
	err = table.ProcessCascades()
	if err != nil {
		g.Log().Error(ctx, err)
		return err
	}
	err = table.ProcessRelatedAndForeign(ctx, genOptions.YamlInputPath, genOptions.GoModuleName, cache)
	if err != nil {
		g.Log().Error(ctx, err)
		return err
	}

	err = doGenCode(ctx, table, genOptions)
	if err != nil {
		g.Log().Error(ctx, err)
		return err
	}
	return nil
}

// 获取生成所需数据
func prepareTemplateData(table *common.TableDef, ctx context.Context) (data g.MapStrStr, err error) {
	//树形菜单选项
	tplData := g.Map{"table": table}
	view := common.TemplateEngine()

	entityKey := "entity"
	entityValue := ""
	var tmpEntity string
	if tmpEntity, err = view.ParseContent(ctx, entityTemplate, tplData); err == nil {
		entityValue = tmpEntity
		entityValue, err = common.TrimBreak(entityValue)
	} else {
		return
	}

	modelKey := "model"
	modelValue := ""
	var tmpModel string
	if tmpModel, err = view.ParseContent(ctx, modelTemplate, tplData); err == nil {
		modelValue = tmpModel
		modelValue, err = common.TrimBreak(modelValue)
	} else {
		return
	}

	daoKey := "dao"
	daoValue := ""
	var tmpDao string
	if tmpDao, err = view.ParseContent(ctx, daoTemplate, tplData); err == nil {
		daoValue = tmpDao
		daoValue, err = common.TrimBreak(daoValue)
	} else {
		return
	}

	daoInternalKey := "dao_internal"
	daoInternalValue := ""
	var tmpInternalDao string
	if tmpInternalDao, err = view.ParseContent(ctx, daoInternalTemplate, tplData); err == nil {
		daoInternalValue = tmpInternalDao
		daoInternalValue, err = common.TrimBreak(daoInternalValue)
	} else {
		return
	}

	controllerKey := "controller"
	controllerValue := ""
	var tmpController string
	if tmpController, err = view.ParseContent(ctx, controllerTemplate, tplData); err == nil {
		controllerValue = tmpController
		controllerValue, err = common.TrimBreak(controllerValue)
	} else {
		return
	}

	serviceKey := "service"
	serviceValue := ""
	var tmpService string
	if tmpService, err = view.ParseContent(ctx, serviceTemplate, tplData); err == nil {
		serviceValue = tmpService
		serviceValue, err = common.TrimBreak(serviceValue)
	} else {
		return
	}

	routerKey := "router"
	routerValue := ""
	var tmpRouter string
	if tmpRouter, err = view.ParseContent(ctx, routerTemplate, tplData); err == nil {
		routerValue = tmpRouter
		routerValue, err = common.TrimBreak(routerValue)
	} else {
		return
	}

	protobufKey := "protobuf"
	protobufValue := ""
	var tmpProtobuf string
	if tmpProtobuf, err = view.ParseContent(ctx, protobufTemplate, tplData); err == nil {
		protobufValue = tmpProtobuf
		protobufValue, err = common.TrimBreak(protobufValue)
	} else {
		return
	}

	providerKey := "provider"
	providerValue := ""
	var tmpProvider string
	if tmpProvider, err = view.ParseContent(ctx, providerTemplate, tplData); err == nil {
		providerValue = tmpProvider
		providerValue, err = common.TrimBreak(providerValue)
	} else {
		return
	}

	sqlKey := "sql"
	sqlValue := ""
	var tmpSql string
	if tmpSql, err = view.ParseContent(ctx, sqlTemplate, tplData); err == nil {
		sqlValue = tmpSql
		sqlValue, err = common.TrimBreak(sqlValue)
	} else {
		return
	}

	jsApiKey := "jsApi"
	jsApiValue := ""
	var tmpJsApi string
	if tmpJsApi, err = view.ParseContent(ctx, jsapiTemplate, tplData); err == nil {
		jsApiValue = tmpJsApi
		jsApiValue, err = common.TrimBreak(jsApiValue)
	} else {
		return
	}

	vueKey := "vue"
	vueValue := ""
	var tmpVue string
	templateContent := listVueTemplate
	if table.TemplateCategory == "tree" {
		//树表
		templateContent = treeVueTemplate
	}
	if tmpVue, err = view.ParseContent(ctx, templateContent, tplData); err == nil {
		vueValue = tmpVue
		vueValue, err = common.TrimBreak(vueValue)
	} else {
		return
	}

	data = g.MapStrStr{
		entityKey:      entityValue,
		modelKey:       modelValue,
		daoKey:         daoValue,
		daoInternalKey: daoInternalValue,
		controllerKey:  controllerValue,
		serviceKey:     serviceValue,
		routerKey:      routerValue,
		protobufKey:    protobufValue,
		providerKey:    providerValue,
		sqlKey:         sqlValue,
		jsApiKey:       jsApiValue,
		vueKey:         vueValue,
	}
	return
}

// 生成代码文件
func doGenCode(ctx context.Context, table *common.TableDef, genOptions *common.GenOptions) error {
	var (
		curDir     string
		path       string
		pbPath     string
		triplePath string
		err        error
	)
	//获取当前运行时目录
	curDir, err = os.Getwd()
	if err != nil {
		return gerror.New("获取本地路径失败")
	}
	frontDir := genOptions.FrontendPath
	if !g.IsEmpty(frontDir) && !gfile.IsDir(frontDir) {
		err = gerror.New("项目前端路径不存在，请检查是否已在配置文件中配置！")
		return err
	}
	var templateData g.MapStrStr
	templateData, err = prepareTemplateData(table, ctx)
	if err != nil {
		return err
	}
	packageName := gstr.TrimLeftStr(table.BackendPackage, genOptions.GoModuleName+"/")
	goFileName := table.GoFileName
	for key, code := range templateData {
		switch key {
		case "controller":
			if genOptions.ServiceOnly {
				break
			}
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/api/", goFileName, ".go"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/api/", goFileName, ".go"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		case "dao":
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/service/internal/dao/", goFileName, ".go"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/service/internal/dao/", goFileName, ".go"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		case "dao_internal":
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/service/internal/dao/internal/", goFileName, ".go"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/service/internal/dao/internal/", goFileName, ".go"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		case "do":
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/service/internal/do/", goFileName, ".go"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/service/internal/do/", goFileName, ".go"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		case "entity":
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/model/entity/", goFileName, ".go"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/model/entity/", goFileName, ".go"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		case "model":
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/model/", goFileName, ".go"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/model/", goFileName, ".go"}, "")
			}
			if !table.IsRpc {
				err = common.WriteFile(path, code, table.Overwrite)
			} else if table.Overwrite {
				if gfile.Exists(path) {
					_ = gfile.Remove(path)
				}
			}
		case "router":
			if genOptions.ServiceOnly {
				break
			}
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/router/", goFileName, ".go"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/router/", goFileName, ".go"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		case "protobuf":
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/proto"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/proto"}, "")
			}
			if table.IsRpc {
				err = common.WriteFile(path+"/"+goFileName+".proto", code, table.Overwrite)
				if err != nil {
					return err
				}
				err = protobuf.CallProtoc(curDir, table.BackendPackage, table.GoFileName, table.SeparatePackage)
				if err != nil {
					return err
				}
			} else if table.Overwrite {
				if gfile.Exists(path) {
					_ = gfile.Remove(path)
				}
				if table.SeparatePackage {
					pbPath = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/model/", goFileName, ".pb.go"}, "")
					triplePath = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/model/", goFileName, "_triple.pb.go"}, "")
				} else {
					pbPath = strings.Join([]string{curDir, "/", packageName, "/model/", goFileName, ".pb.go"}, "")
					triplePath = strings.Join([]string{curDir, "/", packageName, "/model/", goFileName, "_triple.pb.go"}, "")
				}
				if gfile.Exists(pbPath) {
					_ = gfile.Remove(pbPath)
				}
				if gfile.Exists(triplePath) {
					_ = gfile.Remove(triplePath)
				}
			}
		case "provider":
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/provider"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/provider"}, "")
			}
			if table.IsRpc {
				err = common.WriteFile(path+"/"+goFileName+".go", code, table.Overwrite)
			} else if table.Overwrite {
				if gfile.Exists(path) {
					_ = gfile.Remove(path)
				}
			}
		case "service":
			if table.SeparatePackage {
				path = strings.Join([]string{curDir, "/", packageName, "/", goFileName, "/service/", goFileName, ".go"}, "")
			} else {
				path = strings.Join([]string{curDir, "/", packageName, "/service/", goFileName, ".go"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		case "sql":
			if g.IsEmpty(frontDir) {
				break
			}
			path = strings.Join([]string{curDir, "/data/gen_sql/", packageName, "/", goFileName, ".sql"}, "")
			hasSql := gfile.Exists(path)
			err = common.WriteFile(path, code, table.Overwrite)
			if !hasSql || table.Overwrite {
				//第一次生成则向数据库写入菜单数据
				err = saveMenuDb(path, ctx)
				if err != nil {
					return err
				}
			}
		case "vue":
			if g.IsEmpty(frontDir) {
				break
			}
			path = strings.Join([]string{frontDir, "/src/views/", table.FrontendPath, "/", table.FrontendFileName, "/list/index.vue"}, "")
			if gstr.ContainsI(table.BackendPackage, "plugins") {
				path = strings.Join([]string{frontDir, "/src/views/plugins/", table.FrontendPath, "/", table.FrontendFileName, "/list/index.vue"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		case "jsApi":
			if g.IsEmpty(frontDir) {
				break
			}
			path = strings.Join([]string{frontDir, "/src/api/", table.FrontendPath, "/", table.FrontendFileName, ".js"}, "")
			if gstr.ContainsI(table.BackendPackage, "plugins") {
				path = strings.Join([]string{frontDir, "/src/api/plugins/", table.FrontendPath, "/", table.FrontendFileName, ".js"}, "")
			}
			err = common.WriteFile(path, code, table.Overwrite)
		}
	}
	//生成对应的模块路由
	if !genOptions.ServiceOnly {
		err = genModuleRouter(curDir, table.GoFileName, table.BackendPackage, genOptions.GoModuleName, table.Overwrite, table.SeparatePackage)
	}
	return nil
}

// GenModuleRouter 生成模块路由
func genModuleRouter(curDir, goFileName, backendPackage, goModuleName string, overwrite bool, separatePackage bool) (err error) {
	if gstr.CaseSnake(goFileName) == "system" {
		return nil
	}
	packageName := gstr.TrimLeftStr(backendPackage, goModuleName+"/")
	if separatePackage {
		routerFilePath := strings.Join([]string{curDir, "/router/", gstr.Replace(packageName, "/", "_"), "_", goFileName, ".go"}, "")
		if gstr.ContainsI(packageName, "plugins") {
			routerFilePath = strings.Join([]string{curDir, "/plugins/router/", gstr.Replace(packageName, "/", "_"), "_", goFileName, ".go"}, "")
		}
		code := fmt.Sprintf(`package router%simport _ "%s/%s/router"`, "\n", backendPackage, goFileName)
		err = common.WriteFile(routerFilePath, code, overwrite)
	} else {
		routerFilePath := strings.Join([]string{curDir, "/router/", gstr.Replace(packageName, "/", "_"), ".go"}, "")
		if gstr.ContainsI(packageName, "plugins") {
			routerFilePath = strings.Join([]string{curDir, "/plugins/router/", gstr.Replace(packageName, "/", "_"), ".go"}, "")
		}
		code := fmt.Sprintf(`package router%simport _ "%s/router"`, "\n", backendPackage)
		err = common.WriteFile(routerFilePath, code, overwrite)
	}
	return
}

// 写入菜单数据
func saveMenuDb(path string, ctx context.Context) (err error) {
	isAnnotation := false
	var fi *os.File
	fi, err = os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		_ = fi.Close()
	}()
	br := bufio.NewReader(fi)
	var sqlStr []string
	now := gtime.Now()
	var res sql.Result
	var id int64
	var tx *gdb.TX
	tx, err = g.DB(gdb.DefaultGroupName).Ctx(ctx).Begin(ctx)
	if err != nil {
		return
	}
	for {
		bytes, e := br.ReadBytes('\n')
		if e == io.EOF {
			break
		}
		str := gstr.Trim(string(bytes))

		if str == "" {
			continue
		}

		if strings.Contains(str, "/*") {
			isAnnotation = true
		}

		if isAnnotation {
			if strings.Contains(str, "*/") {
				isAnnotation = false
			}
			continue
		}

		if str == "" || strings.HasPrefix(str, "--") || strings.HasPrefix(str, "#") {
			continue
		}
		if strings.HasSuffix(str, ";") {
			if gstr.ContainsI(str, "select") {
				if gstr.ContainsI(str, "@now") {
					continue
				}
				if gstr.ContainsI(str, "@parentId") {
					id, err = res.LastInsertId()
				}
			}
			sqlStr = append(sqlStr, str)
			sqlExec := strings.Join(sqlStr, "")
			gstr.ReplaceByArray(sqlExec, []string{"@parentId", gconv.String(id), "@now", now.Format("Y-m-d H:i:s")})
			//插入业务
			res, err = tx.Exec(sqlExec)
			if err != nil {
				_ = tx.Rollback()
				return
			}
			sqlStr = nil
		} else {
			sqlStr = []string{str}
		}
	}
	_ = tx.Commit()
	return
}
