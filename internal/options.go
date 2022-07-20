package internal

type ImportOptions struct {
	BackendPackage      string
	FrontendModule      string
	GoModuleName        string
	TableNames          []string
	TablePrefixesOnly   []string
	RemoveTablePrefixes []string
	YamlOutputPath      string
	SeparatePackage     bool
	TemplateCategory    string
	Author              string
	Overwrite           bool
	ShowDetail          bool
	IsRpc               bool
}

type GenOptions struct {
	YamlInputPath string
	GoModuleName  string
	ServiceOnly   bool
	FrontendType  string
	FrontendPath  string
}
