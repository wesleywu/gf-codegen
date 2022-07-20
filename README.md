# gf-codegen 代码生成

### 安装 dbimport 和 codegen

```
git clone https://github.com/WesleyWu/gf-codegen.git
cd dbimport
go install
cd ..
cd codegen
go install
```

### 关于代码生成

基于数据库表结构生成完整的前后端CRUD代码。

## 1. 使用代码生成步骤

### 1). 导入表结构
在项目根路径下执行 dbimport，根据表结构自动生成一到多个`{tableName}.yaml`配置文件，每个表对应一个配置文件

命令行参数：
* --dblink 类似 mysql:user:pass@tcp(localhost:3306)/db_name?charset=utf8mb4&parseTime=true&loc=Local 的数据库连接定义
* --tables 指定生成哪些表名的 yaml 文件，多个表名用半角逗号分隔
* --tablePrefixOnly 只需要哪些前缀的表，多个前缀用半角逗号分隔
* --removeTablePrefix 生成时，对应的go文件名需要去掉哪些前缀，多个前缀用半角逗号分隔
* --backendPackage 后端package名称，通常格式是 GoModuleName/Dir1/Dir2/Dir3 或 Dir1/Dir2/Dir3 (可以不指定GoModuleName)，go文件会被放置到相对于项目根路径的 /Dir1/Dir2/Dir3 目录下
* --frontendModule 前端文件所在目录，通常格式是 Dir1/Dir2/Dir3
* --yamlOutputPath 生成的yaml文件放置的目录，缺省为 manifest/config/codegen_conf
* --separatePackage 是否为每一个表都生成单独的业务文件夹，缺省为 false
* --author 业务作者
* --overwrite 下一次生成是否无条件覆盖上次的结果，缺省为 true
* --showDetail 是否生成查看详情前端功能，缺省为 true
* --isRpc 是否生成 DubboGo 方式的 rpc 服务，service为服务提供者（provider），api为服务消费者（consumer），缺省为 false

示例
```
dbimport \
  --dblink="mysql:user:password@tcp(127.0.0.1:3306)/db_name?charset=utf8mb4&parseTime=true&loc=Local" \
  --tables=your_table1,your_table2 \
  --removeTablePrefix=your_ \
  --backendPackage=app/your_package \
  --frontendModule=app/your_package \
  --author=your_name
```

### 2). 编辑配置文件

### 3). 生成代码
在项目根路径下执行 codegen，根据yaml配置文件生成代码

命令行参数：
* --tables 指定生成哪些表名的 yaml 文件，多个表名用半角逗号分隔
* --tablePrefixOnly 只需要哪些前缀的表，多个前缀用半角逗号分隔
* --yamlInputPath yaml配置文件所在路径
* --frontendPath 前端项目在本地硬盘上的根目录
* --frontendType 前端类型，无需指定（目前只支持 arco-design react 前端模板）

示例
```
codegen \
  --serviceOnly=true \
  --tables=your_table1,your_table2
```

## 2. `yaml`配置文件定义
`{tableName}.yaml` 配置文件的定义文档待完善

## 3. 生成代码目录结构（separatePackage=true）
假定：table有两个，表名分别为 `data_book` 和 `data_book_store`，且设定了去掉表前缀 `data_`
### 1). 后端 (Golang) 目录结构

假定：后端 `backendPackage` 为 `app/demo/bookstore`

```html
/app
└── /demo
    └── bookstore
        ├── book
        │   ├── api                     // package api
        │   │   └── book.go             // api(即controller)，处理http请求和返回
        │   ├── model                   // package model
        │   │   ├── entity              // package entity
        │   │   │   └── book.go         // 实体数据模型 struct 定义
        │   │   └── book.go             // 各类输入和输出数据结构 struct 定义
        │   ├── router                  // package router
        │   │   └── book.go             // web请求路由定义
        │   └── service                 // package service
        │       ├── internal            // package internal
        │       │   ├── dao             // package dao
        │       │   │   ├── internal    // package internal
        │       │   │   │   └── book.go // dao 数据访问定义
        │       │   │   └── book.go     // 对 service 暴露 dao 数据访问定义
        │       │   └── do              // package do
        │       │       └── book.go     // 领域对象，用于dao数据操作中业务模型与实例模型转换
        │       └── book.go             // 业务逻辑封装
        └── book_store
            ├── api                
            │   └── book_store.go
            ├── model
            │   ├── entity         
            │   │   └── book_store.go    
            │   └── book_store.go        
            ├── router             
            │   └── book_store.go        
            └── service            
                ├── internal            
                │   ├── dao             
                │   │   ├── internal    
                │   │   │   └── book_store.go 
                │   │   └── book_store.go     
                │   └── do              
                │       └── book_store.go     
                └── book_store.go             
```
注册路由路径为 restful 风格，如下
```html
GET    /demo/bookstore/book/list            // 绑定 api.Book.List 函数，返回数据列表
GET    /demo/bookstore/book/get             // 绑定 api.Book.Get 函数，返回单条数据
POST   /demo/bookstore/book/add             // 绑定 api.Book.Add 函数，处理新增插入
PUT    /demo/bookstore/book/edit            // 绑定 api.Book.Edit 函数，处理编辑更新
DELETE /demo/bookstore/book/delete          // 绑定 api.Book.Delete 函数，处理单条及多条数据删除
PUT    /demo/bookstore/book/changeXXX       // 如果结果表格支持inline编辑并更新特定字段

GET    /demo/bookstore/book-store/list      // 绑定 api.BookStore.List 函数，返回数据列表
GET    /demo/bookstore/book-store/get       // 绑定 api.BookStore.Get 函数，返回单条数据
POST   /demo/bookstore/book-store/add       // 绑定 api.BookStore.Add 函数，处理新增插入
PUT    /demo/bookstore/book-store/edit      // 绑定 api.BookStore.Edit 函数，处理编辑更新
DELETE /demo/bookstore/book-store/delete    // 绑定 api.BookStore.Delete 函数，处理单条及多条数据删除
PUT    /demo/bookstore/book-store/changeXXX // 如果结果表格支持inline编辑并更新特定字段
```
### 2). 前端 (Vue.js) 目录结构
假定前端模块 `frontendModule` 为 `demo/bookstore`
```html
/src
├── /api
│   └── demo
│       └── bookstore
│           ├── book.js           // 处理到后端 api http请求的函数
│           └── book-store.js     // 处理到后端 api http请求的函数
└── views
    └── demo
        └── bookstore
            ├── book
            │   └── list
            │       └── index.vue // Vue组件 - CRUD 界面
            └── book-store
                └── list
                    └── index.vue // Vue组件 - CRUD 界面
```
组件访问路径分别为
```html
demo/bookstore/book/list
demo/bookstore/book-store/list
```

## 4. 命名规范




