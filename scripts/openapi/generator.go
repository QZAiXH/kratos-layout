package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	// openAPIReleaseVersion 表示仓库发布使用的 OpenAPI 版本号。
	openAPIReleaseVersion = "3.1.0"
	// openAPIBaselineVersion 表示 gnostic 当前生成的基线版本号。
	openAPIBaselineVersion = "3.0.3"
	// openAPIJSONSchemaDialect 表示发布产物要求的根级 JSON Schema 方言。
	openAPIJSONSchemaDialect = "https://spec.openapis.org/oas/3.1/dialect/base"
	// openAPIPublisherExtension 表示发布器写入的 OpenAPI 扩展字段名。
	openAPIPublisherExtension = "x-openapi-publisher"
	// openAPIPublisherFormat 表示发布器格式标识。
	openAPIPublisherFormat = "kratos-openapi-publisher"
)

// generator 负责从 proto 生成模块文档与 bundle。
type generator struct {
	rootDir string // 仓库根目录

	moduleOutputDir string // 模块文档输出目录

	overlayDir string // overlay 目录

	bundleOutputPath string // bundle 输出路径

	bufBin string // buf 可执行文件路径
}

// moduleSpec 描述单个模块的发现结果与输出约定。
type moduleSpec struct {
	Domain string // 一级域目录，例如 user

	API string // 二级 API 目录，例如 admin

	Version string // 版本目录，例如 v1

	Dir string // 模块目录的绝对路径

	ProtoFiles []string // 参与模块文档生成的 proto 相对路径
}

// Name 返回模块输出名。
func (m moduleSpec) Name() string {
	if m.API == "" {
		return fmt.Sprintf("%s_%s", m.Domain, m.Version)
	}
	return fmt.Sprintf("%s_%s_%s", m.Domain, m.API, m.Version)
}

// OutputPath 返回模块产物输出路径。
func (m moduleSpec) OutputPath(moduleOutputDir string) string {
	return filepath.Join(moduleOutputDir, m.Name()+".openapi.json")
}

// Title 返回模块文档标题。
func (m moduleSpec) Title() string {
	parts := []string{titleWord(m.Domain)}
	if m.API != "" {
		parts = append(parts, titleWord(m.API))
	}
	parts = append(parts, strings.ToUpper(m.Version), "API")
	return strings.Join(parts, " ")
}

// Description 返回模块文档描述。
func (m moduleSpec) Description() string {
	return fmt.Sprintf("%s 模块接口文档。", strings.Join(m.pathParts(), "/"))
}

// pathParts 返回模块相对 api 根目录的路径片段。
func (m moduleSpec) pathParts() []string {
	parts := []string{m.Domain}
	if m.API != "" {
		parts = append(parts, m.API)
	}
	return append(parts, m.Version)
}

// bundleResult 描述一次生成流程的最终输出。
type bundleResult struct {
	Modules []moduleSpec // 已生成的模块集合

	ModuleOutputDir string // 模块输出目录

	BundleOutputPath string // bundle 输出路径
}

// newGenerator 组装生成器配置。
func newGenerator(cfg cliConfig) (*generator, error) {
	rootDir, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve root dir: %w", err)
	}
	moduleOutputDir := filepath.Join(rootDir, cfg.ModuleOutputDir)
	overlayDir := filepath.Join(rootDir, cfg.OverlayDir)
	bundleOutputPath := filepath.Join(rootDir, cfg.BundleOutputPath)
	return &generator{
		rootDir:          rootDir,
		moduleOutputDir:  moduleOutputDir,
		overlayDir:       overlayDir,
		bundleOutputPath: bundleOutputPath,
		bufBin:           cfg.BufBin,
	}, nil
}

// Generate 发现模块、生成模块文档并聚合 bundle。
func (g *generator) Generate(ctx context.Context) (*bundleResult, error) {
	if err := g.failOnLegacySwagger(); err != nil {
		return nil, err
	}
	modules, err := discoverModules(g.rootDir)
	if err != nil {
		return nil, err
	}
	if len(modules) == 0 {
		return nil, fmt.Errorf("no api modules discovered under %s", filepath.Join(g.rootDir, "api"))
	}
	if err := g.prepareOutputDirs(); err != nil {
		return nil, err
	}

	publishedDocs := make([]publishedDocument, 0, len(modules))
	for _, module := range modules {
		baseline, comments, err := g.generateBaseline(ctx, module)
		if err != nil {
			return nil, err
		}
		overlay, err := loadOptionalYAMLMap(filepath.Join(g.overlayDir, module.Name()+".yaml"))
		if err != nil {
			return nil, fmt.Errorf("load overlay for %s: %w", module.Name(), err)
		}
		doc, output, err := publishModuleDocument(module, baseline, overlay, comments)
		if err != nil {
			return nil, err
		}
		outputPath := module.OutputPath(g.moduleOutputDir)
		if err := os.WriteFile(outputPath, output, 0o644); err != nil {
			return nil, fmt.Errorf("write module document %s: %w", outputPath, err)
		}
		publishedDocs = append(publishedDocs, publishedDocument{Module: module, Document: doc})
	}

	bundleDoc, bundleBytes, err := buildBundleDocument(publishedDocs)
	if err != nil {
		return nil, err
	}
	if err := validatePublishedDocument(bundleDoc, publicationTarget{Name: "openapi_bundle", Kind: publicationKindBundle}); err != nil {
		return nil, fmt.Errorf("validate bundle document: %w", err)
	}
	if err := os.WriteFile(g.bundleOutputPath, bundleBytes, 0o644); err != nil {
		return nil, fmt.Errorf("write bundle document %s: %w", g.bundleOutputPath, err)
	}

	return &bundleResult{
		Modules:          modules,
		ModuleOutputDir:  g.moduleOutputDir,
		BundleOutputPath: g.bundleOutputPath,
	}, nil
}

// failOnLegacySwagger 在仓库仍存在旧 Swagger 产物时直接失败。
func (g *generator) failOnLegacySwagger() error {
	legacy := []string{}
	apiRoot := filepath.Join(g.rootDir, "api")
	err := filepath.WalkDir(apiRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".swagger.json") {
			return nil
		}
		legacy = append(legacy, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("scan legacy swagger files: %w", err)
	}
	sort.Strings(legacy)
	if len(legacy) == 0 {
		return nil
	}
	return fmt.Errorf("legacy swagger artifacts still exist: %s", strings.Join(legacy, ", "))
}

// prepareOutputDirs 清理旧模块产物并确保目录存在。
func (g *generator) prepareOutputDirs() error {
	for _, dir := range []string{g.moduleOutputDir, filepath.Dir(g.bundleOutputPath)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	if err := removeMatchingFiles(g.moduleOutputDir, "*.openapi.json"); err != nil {
		return err
	}
	if err := removeMatchingFiles(filepath.Dir(g.bundleOutputPath), "*.openapi.json"); err != nil {
		return err
	}
	return nil
}

// generateBaseline 调用 protoc-gen-openapi 为单个模块生成 YAML 基线，并提取 proto 注释索引。
func (g *generator) generateBaseline(ctx context.Context, module moduleSpec) ([]byte, *protoCommentIndex, error) {
	tmpDir, err := os.MkdirTemp("", "kratos-openapi-*")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp dir for %s: %w", module.Name(), err)
	}
	defer os.RemoveAll(tmpDir)

	descriptorSetPath := filepath.Join(tmpDir, "module.pb")
	if err := g.buildDescriptorSet(ctx, module, descriptorSetPath); err != nil {
		return nil, nil, err
	}
	if err := g.generateOpenAPIBaseline(ctx, module, tmpDir); err != nil {
		return nil, nil, err
	}
	baselinePath := filepath.Join(tmpDir, "openapi.yaml")
	content, err := os.ReadFile(baselinePath)
	if err != nil {
		return nil, nil, fmt.Errorf("read baseline %s: %w", baselinePath, err)
	}
	comments, err := loadProtoCommentIndex(module, descriptorSetPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load proto comment index for %s: %w", module.Name(), err)
	}
	return content, comments, nil
}

// buildDescriptorSet 使用 buf build 输出包含 source info 的 descriptor set。
func (g *generator) buildDescriptorSet(ctx context.Context, module moduleSpec, outputPath string) error {
	args := []string{"build", ".", "--as-file-descriptor-set", "--output=" + outputPath}
	for _, protoFile := range module.ProtoFiles {
		args = append(args, "--path="+protoFile)
	}
	if err := g.runBuf(ctx, args...); err != nil {
		return fmt.Errorf("build descriptor set for %s: %w", module.Name(), err)
	}
	return nil
}

// generateOpenAPIBaseline 使用 buf generate 调用 protoc-gen-openapi 生成模块 YAML 基线。
func (g *generator) generateOpenAPIBaseline(ctx context.Context, module moduleSpec, outputDir string) error {
	templatePath := filepath.Join(outputDir, "buf.gen.openapi.yaml")
	template := map[string]any{
		"version": "v2",
		"plugins": []any{
			map[string]any{
				"local":    []string{"go", "run", "github.com/google/gnostic/cmd/protoc-gen-openapi@v0.7.1"},
				"out":      outputDir,
				"strategy": "all",
				"opt": []string{
					"output_mode=merged",
					"naming=proto",
					"enum_type=string",
					"fq_schema_naming=true",
					"default_response=true",
				},
			},
		},
	}
	content, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("marshal openapi buf template for %s: %w", module.Name(), err)
	}
	if err := os.WriteFile(templatePath, content, 0o644); err != nil {
		return fmt.Errorf("write openapi buf template for %s: %w", module.Name(), err)
	}

	args := []string{"generate", ".", "--template=" + templatePath}
	for _, protoFile := range module.ProtoFiles {
		args = append(args, "--path="+protoFile)
	}
	if err := g.runBuf(ctx, args...); err != nil {
		return fmt.Errorf("generate baseline for %s: %w", module.Name(), err)
	}
	return nil
}

func (g *generator) runBuf(ctx context.Context, args ...string) error {
	cmd := g.bufCommand(ctx, args...)
	cmd.Dir = g.rootDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (g *generator) bufCommand(ctx context.Context, args ...string) *exec.Cmd {
	if g.bufBin != "" {
		if path, err := exec.LookPath(g.bufBin); err == nil {
			return exec.CommandContext(ctx, path, args...)
		}
	}
	goArgs := append([]string{"run", "github.com/bufbuild/buf/cmd/buf@latest"}, args...)
	return exec.CommandContext(ctx, "go", goArgs...)
}

// discoverModules 发现所有 api/<domain>/<version> 与 api/<domain>/<api>/<version> 模块。
func discoverModules(rootDir string) ([]moduleSpec, error) {
	apiRoot := filepath.Join(rootDir, "api")
	modules := make([]moduleSpec, 0)
	err := filepath.WalkDir(apiRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if path == apiRoot || !strings.HasPrefix(entry.Name(), "v") {
			return nil
		}
		relPath, err := filepath.Rel(apiRoot, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		if len(parts) != 2 && len(parts) != 3 {
			return filepath.SkipDir
		}
		protoFiles, err := listModuleProtoFiles(rootDir, path)
		if err != nil {
			return err
		}
		if len(protoFiles) == 0 {
			return filepath.SkipDir
		}
		module := moduleSpec{
			Domain:     parts[0],
			Version:    parts[len(parts)-1],
			Dir:        path,
			ProtoFiles: protoFiles,
		}
		if len(parts) == 3 {
			module.API = parts[1]
		}
		modules = append(modules, module)
		return filepath.SkipDir
	})
	if err != nil {
		return nil, fmt.Errorf("discover api modules: %w", err)
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name() < modules[j].Name()
	})
	return modules, nil
}

// listModuleProtoFiles 返回模块参与文档生成的 proto 文件列表。
func listModuleProtoFiles(rootDir, moduleDir string) ([]string, error) {
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("read module dir %s: %w", moduleDir, err)
	}
	protoFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".proto") || strings.HasSuffix(name, "_error.proto") {
			continue
		}
		relPath, err := filepath.Rel(rootDir, filepath.Join(moduleDir, name))
		if err != nil {
			return nil, fmt.Errorf("build relative path for %s: %w", name, err)
		}
		protoFiles = append(protoFiles, filepath.ToSlash(relPath))
	}
	sort.Strings(protoFiles)
	return protoFiles, nil
}

// loadOptionalYAMLMap 在 overlay 文件存在时解析为 map，否则返回 nil。
func loadOptionalYAMLMap(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var raw any
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	if raw == nil {
		return nil, nil
	}
	mapped, ok := normalizeYAMLValue(raw).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("overlay root must be an object")
	}
	return mapped, nil
}

// removeMatchingFiles 删除目录下匹配 glob 的文件。
func removeMatchingFiles(dir, pattern string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read output dir %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matched, err := filepath.Match(pattern, entry.Name())
		if err != nil {
			return fmt.Errorf("match pattern %s against %s: %w", pattern, entry.Name(), err)
		}
		if !matched {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", filepath.Join(dir, entry.Name()), err)
		}
	}
	return nil
}

// normalizeYAMLValue 把 yaml 解码后的动态值统一转换成 JSON 友好的结构。
func normalizeYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = normalizeYAMLValue(item)
		}
		return result
	case map[any]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[fmt.Sprint(key)] = normalizeYAMLValue(item)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, normalizeYAMLValue(item))
		}
		return result
	default:
		return value
	}
}

// marshalCanonicalJSON 以稳定格式输出 JSON。
func marshalCanonicalJSON(value any) ([]byte, error) {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(content, '\n'), nil
}

// publishedDocument 表示已经通过发布校验的模块文档。
type publishedDocument struct {
	Module moduleSpec // 文档所属模块

	Document map[string]any // 模块文档内容
}

// publicationKind 表示发布目标类型。
type publicationKind string

const (
	// publicationKindModule 表示模块文档。
	publicationKindModule publicationKind = "module"
	// publicationKindBundle 表示聚合 bundle。
	publicationKindBundle publicationKind = "bundle"
)

// publicationTarget 描述发布校验时所需的目标元信息。
type publicationTarget struct {
	Name string // 发布目标名称

	Kind publicationKind // 发布目标类型
}

// copyMap 深拷贝动态 map，避免 merge 时污染原始输入。
func copyMap(value map[string]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = cloneValue(item)
	}
	return result
}

// cloneValue 深拷贝动态值。
func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return copyMap(typed)
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, cloneValue(item))
		}
		return result
	default:
		return typed
	}
}

// titleWord 以最小规则把目录名转换为标题词。
func titleWord(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

// walkMaps 递归遍历 map/array 中的所有节点。
func walkMaps(value any, fn func(path []string, current any) error) error {
	return walkMapsAtPath(nil, value, fn)
}

// walkMapsAtPath 递归遍历动态值，并保留当前访问路径。
func walkMapsAtPath(path []string, value any, fn func(path []string, current any) error) error {
	if err := fn(path, value); err != nil {
		return err
	}
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := walkMapsAtPath(append(path, key), typed[key], fn); err != nil {
				return err
			}
		}
	case []any:
		for idx, item := range typed {
			if err := walkMapsAtPath(append(path, fmt.Sprintf("%d", idx)), item, fn); err != nil {
				return err
			}
		}
	}
	return nil
}
