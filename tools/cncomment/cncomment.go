package cncomment

import (
	"go/ast"
	"unicode"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

// init 注册中文注释检查插件。
func init() {
	register.Plugin("cncomment", newPlugin)
}

// plugin 承载 golangci-lint module plugin 接口。
type plugin struct{}

// newPlugin 创建中文注释检查插件实例。
func newPlugin(any) (register.LinterPlugin, error) {
	return plugin{}, nil
}

// BuildAnalyzers 返回当前插件提供的 analyzer。
func (plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{Analyzer}, nil
}

// GetLoadMode 声明当前规则只需要语法树。
func (plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

// Analyzer 检查导出 Go API 是否有中文注释。
var Analyzer = &analysis.Analyzer{
	Name: "cncomment",
	Doc:  "check exported Go declarations have Chinese comments",
	Run:  run,
}

// run 遍历当前 package 的手写 Go 文件并检查导出声明注释。
func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				checkFunc(pass, d)
			case *ast.GenDecl:
				checkTypes(pass, d)
			}
		}
	}
	return nil, nil
}

// checkFunc 检查导出函数和方法的中文注释。
func checkFunc(pass *analysis.Pass, decl *ast.FuncDecl) {
	if !decl.Name.IsExported() || hasChineseComment(decl.Doc) {
		return
	}
	if decl.Recv != nil {
		pass.Reportf(decl.Pos(), "导出方法缺少中文注释")
		return
	}
	pass.Reportf(decl.Pos(), "导出函数缺少中文注释")
}

// checkTypes 检查导出结构体、接口及其成员注释。
func checkTypes(pass *analysis.Pass, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		switch typ := typeSpec.Type.(type) {
		case *ast.StructType:
			if typeSpec.Name.IsExported() && !hasChineseTypeComment(decl, typeSpec) {
				pass.Reportf(typeSpec.Pos(), "导出结构体缺少中文注释")
			}
			checkStructFields(pass, typ)
		case *ast.InterfaceType:
			if typeSpec.Name.IsExported() && !hasChineseTypeComment(decl, typeSpec) {
				pass.Reportf(typeSpec.Pos(), "导出接口缺少中文注释")
			}
			checkInterfaceMethods(pass, typ)
		}
	}
}

// checkStructFields 检查首字母大写的结构体字段中文注释。
func checkStructFields(pass *analysis.Pass, typ *ast.StructType) {
	for _, field := range typ.Fields.List {
		if !hasExportedField(field) || hasChineseFieldComment(field) {
			continue
		}
		pass.Reportf(field.Pos(), "导出结构体字段缺少中文注释")
	}
}

// checkInterfaceMethods 检查首字母大写的接口方法中文注释。
func checkInterfaceMethods(pass *analysis.Pass, typ *ast.InterfaceType) {
	for _, method := range typ.Methods.List {
		if !hasExportedField(method) || hasChineseFieldComment(method) {
			continue
		}
		pass.Reportf(method.Pos(), "导出接口方法缺少中文注释")
	}
}

// hasExportedField 判断字段列表里是否存在首字母大写的名字。
func hasExportedField(field *ast.Field) bool {
	if len(field.Names) == 0 {
		return exportedExpr(field.Type)
	}
	for _, name := range field.Names {
		if name.IsExported() {
			return true
		}
	}
	return false
}

// exportedExpr 判断匿名字段类型是否是导出标识符。
func exportedExpr(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.IsExported()
	case *ast.SelectorExpr:
		return e.Sel.IsExported()
	case *ast.StarExpr:
		return exportedExpr(e.X)
	default:
		return false
	}
}

// hasChineseTypeComment 判断类型声明是否有中文注释。
func hasChineseTypeComment(decl *ast.GenDecl, spec *ast.TypeSpec) bool {
	return hasChineseComment(spec.Doc) || hasChineseComment(decl.Doc) || hasChineseComment(spec.Comment)
}

// hasChineseFieldComment 判断字段或接口方法是否有中文注释。
func hasChineseFieldComment(field *ast.Field) bool {
	return hasChineseComment(field.Doc) || hasChineseComment(field.Comment)
}

// hasChineseComment 判断注释组是否包含汉字。
func hasChineseComment(group *ast.CommentGroup) bool {
	if group == nil {
		return false
	}
	for _, comment := range group.List {
		for _, r := range comment.Text {
			if unicode.Is(unicode.Han, r) {
				return true
			}
		}
	}
	return false
}
