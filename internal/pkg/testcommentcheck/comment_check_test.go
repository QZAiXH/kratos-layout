package testcommentcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

var (
	testFuncPattern = regexp.MustCompile(`^func\s+(Test\w+)\s*\(`)
	runPattern      = regexp.MustCompile(`^\s*t\.Run\(`)
	hanPattern      = regexp.MustCompile(`[\p{Han}]`)
)

// finding 描述单个测试注释违规的位置和原因。
type finding struct {
	Path    string
	Line    int
	Message string
}

// String 返回适合直接打印到失败信息中的违规描述。
func (f finding) String() string {
	return fmt.Sprintf("%s:%d: %s", f.Path, f.Line, f.Message)
}

// TestRepositoryTestComments 校验仓库内所有 Go 测试都带有符合规范的中文注释。
func TestRepositoryTestComments(t *testing.T) {
	root := locateRepoRoot(t)
	files := collectTestFiles(t, root)

	var findings []finding
	for _, file := range files {
		findings = append(findings, inspectTestFile(t, root, file)...)
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

	t.Fatalf("发现测试注释不符合规范的用例:\n%s", strings.Join(messages, "\n"))
}

// TestInspectTestSource 覆盖守卫对顶层测试、子用例和非测试函数的判定边界。
func TestInspectTestSource(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []finding
	}{
		{
			name: "reports missing top-level comment",
			src: strings.Join([]string{
				"package sample",
				"",
				"func TestMissing(t *testing.T) {",
				"}",
				"",
			}, "\n"),
			want: []finding{{
				Path:    "sample_test.go",
				Line:    3,
				Message: "顶层测试缺少紧邻的中文前置注释",
			}},
		},
		{
			name: "reports english-only top-level comment",
			src: strings.Join([]string{
				"package sample",
				"",
				"// TestMissing validates top-level behavior.",
				"func TestMissing(t *testing.T) {",
				"}",
				"",
			}, "\n"),
			want: []finding{{
				Path:    "sample_test.go",
				Line:    4,
				Message: "顶层测试缺少紧邻的中文前置注释",
			}},
		},
		{
			name: "reports missing run comment",
			src: strings.Join([]string{
				"package sample",
				"",
				"// TestHasRun 验证顶层测试已有中文注释。",
				"func TestHasRun(t *testing.T) {",
				"\tt.Run(\"case\", func(t *testing.T) {",
				"\t})",
				"}",
				"",
			}, "\n"),
			want: []finding{{
				Path:    "sample_test.go",
				Line:    5,
				Message: "t.Run 子用例缺少中文场景注释",
			}},
		},
		{
			name: "accepts top-level and body comments",
			src: strings.Join([]string{
				"package sample",
				"",
				"// TestHasRun 验证顶层测试已有中文注释。",
				"func TestHasRun(t *testing.T) {",
				"\tt.Run(\"case\", func(t *testing.T) {",
				"\t\t// 子用例说明这个分支会通过守卫。",
				"\t})",
				"}",
				"",
			}, "\n"),
			want: nil,
		},
		{
			name: "ignores helper functions",
			src: strings.Join([]string{
				"package sample",
				"",
				"func helper(t *testing.T) {",
				"}",
				"",
			}, "\n"),
			want: nil,
		},
	}

	for _, tc := range cases {
		got := inspectTestSource("sample_test.go", tc.src)
		if tc.want == nil {
			if len(got) != 0 {
				t.Fatalf("%s: inspectTestSource() = %#v, want empty", tc.name, got)
			}
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%s: inspectTestSource() = %#v, want %#v", tc.name, got, tc.want)
		}
	}
}

// locateRepoRoot 从当前测试文件位置向上查找仓库根目录。
func locateRepoRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve current test file")
	}

	dir := filepath.Dir(currentFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("failed to locate repository root from %s", currentFile)
		}
		dir = parent
	}
}

// collectTestFiles 收集仓库中需要参与守卫检查的 `*_test.go` 文件。
func collectTestFiles(t *testing.T, root string) []string {
	t.Helper()

	files := make([]string, 0, 32)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path == root {
				return nil
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "third_party" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(d.Name(), "_test.go") {
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

// inspectTestFile 检查单个测试文件里的顶层测试和 `t.Run` 子用例注释。
func inspectTestFile(t *testing.T, root, path string) []finding {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	relPath, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}

	return inspectTestSource(filepath.ToSlash(relPath), string(content))
}

// inspectTestSource 基于源码文本提取测试注释违规项。
func inspectTestSource(path, src string) []finding {
	lines := strings.Split(src, "\n")
	findings := make([]finding, 0)

	for idx, line := range lines {
		if testFuncPattern.MatchString(line) && !hasChineseCommentAbove(lines, idx) {
			findings = append(findings, finding{
				Path:    path,
				Line:    idx + 1,
				Message: "顶层测试缺少紧邻的中文前置注释",
			})
		}

		if runPattern.MatchString(line) && !hasChineseRunComment(lines, idx) {
			findings = append(findings, finding{
				Path:    path,
				Line:    idx + 1,
				Message: "t.Run 子用例缺少中文场景注释",
			})
		}
	}

	return findings
}

// hasChineseRunComment 判断 `t.Run` 是否具备前置注释或子用例体开头注释。
func hasChineseRunComment(lines []string, idx int) bool {
	return hasChineseCommentAbove(lines, idx) || hasChineseCommentAtBodyStart(lines, idx+1)
}

// hasChineseCommentAbove 判断目标行上方是否紧邻连续的中文 `//` 注释块。
func hasChineseCommentAbove(lines []string, idx int) bool {
	if idx == 0 {
		return false
	}

	if !isLineComment(lines[idx-1]) {
		return false
	}

	start := idx - 1
	for start > 0 && isLineComment(lines[start-1]) {
		start--
	}

	return blockHasChinese(lines[start:idx])
}

// hasChineseCommentAtBodyStart 判断子用例函数体开头的第一段注释是否包含中文。
func hasChineseCommentAtBodyStart(lines []string, idx int) bool {
	for idx < len(lines) && strings.TrimSpace(lines[idx]) == "" {
		idx++
	}
	if idx >= len(lines) || !isLineComment(lines[idx]) {
		return false
	}

	end := idx
	for end < len(lines) && isLineComment(lines[end]) {
		end++
	}

	return blockHasChinese(lines[idx:end])
}

// isLineComment 判断一行是否为 `//` 注释。
func isLineComment(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "//")
}

// blockHasChinese 判断注释块里是否包含中文字符。
func blockHasChinese(lines []string) bool {
	return hanPattern.MatchString(strings.Join(lines, "\n"))
}
