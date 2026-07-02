package commentcheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var hanPattern = regexp.MustCompile(`[\p{Han}]`)

// finding 描述单个中文注释违规位置。
type finding struct {
	Path    string // Path 是相对仓库根目录的文件路径。
	Line    int    // Line 是违规声明所在行号。
	Message string // Message 是面向修复者的违规说明。
}

// String 返回适合测试失败输出的违规描述。
func (f finding) String() string {
	return fmt.Sprintf("%s:%d: %s", f.Path, f.Line, f.Message)
}

// TestGoChineseComments 校验手写 Go 代码中的声明都有中文注释。
func TestGoChineseComments(t *testing.T) {
	root := locateRepoRoot(t)
	files := collectGoFiles(t, root)

	var findings []finding
	for _, file := range files {
		findings = append(findings, inspectGoFile(t, root, file)...)
	}
	if len(findings) == 0 {
		return
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path == findings[j].Path {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Path < findings[j].Path
	})

	messages := make([]string, 0, len(findings))
	for _, item := range findings {
		messages = append(messages, item.String())
	}
	t.Fatalf("发现 Go 中文注释不符合规范的声明:\n%s", strings.Join(messages, "\n"))
}

// locateRepoRoot 从当前测试文件向上定位仓库根目录。
func locateRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("failed to locate repository root from %s", dir)
		}
		dir = parent
	}
}

// collectGoFiles 收集需要检查的手写 Go 文件。
func collectGoFiles(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldSkipDir(root, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".go") && !shouldSkipFile(root, path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	return files
}

// shouldSkipDir 判断目录是否属于生成物或工具缓存。
func shouldSkipDir(root, path string) bool {
	if path == root {
		return false
	}
	rel := filepath.ToSlash(mustRel(root, path))
	name := filepath.Base(path)
	if strings.HasPrefix(name, ".") || name == "bin" || name == "third_party" {
		return true
	}
	if strings.HasPrefix(rel, "internal/data/ent/") && !strings.HasPrefix(rel, "internal/data/ent/schema") {
		return true
	}
	return false
}

// shouldSkipFile 判断文件是否属于生成代码。
func shouldSkipFile(root, path string) bool {
	rel := filepath.ToSlash(mustRel(root, path))
	name := filepath.Base(path)
	if strings.HasSuffix(name, ".pb.go") || strings.HasSuffix(name, "_grpc.pb.go") || strings.HasSuffix(name, "_http.pb.go") {
		return true
	}
	if name == "wire_gen.go" {
		return true
	}
	if strings.HasPrefix(rel, "internal/data/ent/") && !strings.HasPrefix(rel, "internal/data/ent/schema/") && rel != "internal/data/ent/sql_driver.go" {
		return true
	}
	return false
}

// inspectGoFile 检查单个 Go 文件中的函数、方法、结构体、字段、接口和接口方法注释。
func inspectGoFile(t *testing.T, root, path string) []finding {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	rel := filepath.ToSlash(mustRel(root, path))
	var findings []finding
	for _, decl := range file.Decls {
		switch typed := decl.(type) {
		case *ast.FuncDecl:
			findings = append(findings, inspectFuncDecl(fset, rel, typed)...)
		case *ast.GenDecl:
			findings = append(findings, inspectGenDecl(fset, rel, typed)...)
		}
	}
	return findings
}

// inspectFuncDecl 检查函数或方法声明注释。
func inspectFuncDecl(fset *token.FileSet, path string, decl *ast.FuncDecl) []finding {
	if hasChineseComment(decl.Doc) {
		return nil
	}
	kind := "函数"
	if decl.Recv != nil {
		kind = "方法"
	}
	return []finding{newFinding(fset, path, decl.Pos(), kind+"缺少中文注释")}
}

// inspectGenDecl 检查类型声明中的结构体与接口注释。
func inspectGenDecl(fset *token.FileSet, path string, decl *ast.GenDecl) []finding {
	var findings []finding
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		switch typed := typeSpec.Type.(type) {
		case *ast.StructType:
			if !hasChineseTypeComment(decl, typeSpec) {
				findings = append(findings, newFinding(fset, path, typeSpec.Pos(), "结构体缺少中文注释"))
			}
			findings = append(findings, inspectStructFields(fset, path, typed)...)
		case *ast.InterfaceType:
			if !hasChineseTypeComment(decl, typeSpec) {
				findings = append(findings, newFinding(fset, path, typeSpec.Pos(), "接口缺少中文注释"))
			}
			findings = append(findings, inspectInterfaceMethods(fset, path, typed)...)
		}
	}
	return findings
}

// inspectStructFields 检查结构体每个字段的中文注释。
func inspectStructFields(fset *token.FileSet, path string, typ *ast.StructType) []finding {
	var findings []finding
	for _, field := range typ.Fields.List {
		if hasChineseFieldComment(field) {
			continue
		}
		findings = append(findings, newFinding(fset, path, field.Pos(), "结构体字段缺少中文注释"))
	}
	return findings
}

// inspectInterfaceMethods 检查接口方法或嵌入接口的中文注释。
func inspectInterfaceMethods(fset *token.FileSet, path string, typ *ast.InterfaceType) []finding {
	var findings []finding
	for _, method := range typ.Methods.List {
		if hasChineseFieldComment(method) {
			continue
		}
		findings = append(findings, newFinding(fset, path, method.Pos(), "接口方法缺少中文注释"))
	}
	return findings
}

// hasChineseTypeComment 判断类型声明是否有中文注释。
func hasChineseTypeComment(decl *ast.GenDecl, spec *ast.TypeSpec) bool {
	return hasChineseComment(spec.Doc) || hasChineseComment(decl.Doc) || hasChineseComment(spec.Comment)
}

// hasChineseFieldComment 判断字段或接口方法是否有中文注释。
func hasChineseFieldComment(field *ast.Field) bool {
	return hasChineseComment(field.Doc) || hasChineseComment(field.Comment)
}

// hasChineseComment 判断注释组是否包含中文字符。
func hasChineseComment(group *ast.CommentGroup) bool {
	return group != nil && hanPattern.MatchString(group.Text())
}

// newFinding 根据 token 位置生成违规项。
func newFinding(fset *token.FileSet, path string, pos token.Pos, message string) finding {
	position := fset.Position(pos)
	return finding{Path: path, Line: position.Line, Message: message}
}

// mustRel 返回相对路径，失败时退回原始路径。
func mustRel(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}
