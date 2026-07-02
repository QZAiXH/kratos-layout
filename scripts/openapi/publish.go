package main

import (
	"fmt"
	"sort"
	"strings"
)

// publishModuleDocument 把 gnostic 基线、proto 注释与 overlay 编排成最终模块文档。
func publishModuleDocument(module moduleSpec, baseline []byte, overlay map[string]any, comments *protoCommentIndex) (map[string]any, []byte, error) {
	baselineDoc, err := parseDocumentMap(baseline)
	if err != nil {
		return nil, nil, fmt.Errorf("parse baseline for %s: %w", module.Name(), err)
	}
	published := copyMap(baselineDoc)
	enrichDocumentWithProtoComments(published, comments)
	if overlay != nil {
		published, err = mergeDocument(published, overlay)
		if err != nil {
			return nil, nil, fmt.Errorf("merge overlay for %s: %w", module.Name(), err)
		}
	}
	normalizeModuleDocument(published, module, comments)
	if err := validatePublishedDocument(published, publicationTarget{Name: module.Name(), Kind: publicationKindModule}); err != nil {
		return nil, nil, fmt.Errorf("validate module %s: %w", module.Name(), err)
	}
	output, err := marshalCanonicalJSON(published)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal module %s: %w", module.Name(), err)
	}
	return published, output, nil
}

// parseDocumentMap 把 YAML/JSON OpenAPI 文档解析为动态 map。
func parseDocumentMap(content []byte) (map[string]any, error) {
	var raw any
	if err := unmarshalOpenAPI(content, &raw); err != nil {
		return nil, err
	}
	mapped, ok := normalizeYAMLValue(raw).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("openapi document root must be an object")
	}
	return mapped, nil
}

// mergeDocument 以 merge patch 语义应用 overlay。
func mergeDocument(base map[string]any, overlay map[string]any) (map[string]any, error) {
	merged, ok := mergeValue(base, overlay).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("merged document root must remain an object")
	}
	return merged, nil
}

// mergeValue 递归合并动态值，null 表示删除。
func mergeValue(base, overlay any) any {
	if overlay == nil {
		return nil
	}
	baseMap, baseIsMap := base.(map[string]any)
	overlayMap, overlayIsMap := overlay.(map[string]any)
	if !baseIsMap || !overlayIsMap {
		return cloneValue(overlay)
	}

	result := copyMap(baseMap)
	for key, value := range overlayMap {
		if value == nil {
			delete(result, key)
			continue
		}
		existing, exists := result[key]
		if exists {
			result[key] = mergeValue(existing, value)
			continue
		}
		result[key] = cloneValue(value)
	}
	return result
}

// normalizeModuleDocument 补齐发布契约元信息并统一模块 tag。
func normalizeModuleDocument(doc map[string]any, module moduleSpec, comments *protoCommentIndex) {
	info := ensureMap(doc, "info")
	setDefaultString(info, "title", module.Title())
	setDefaultString(info, "version", module.Version)
	setDefaultString(info, "description", module.Description())

	doc["openapi"] = openAPIReleaseVersion
	doc["jsonSchemaDialect"] = openAPIJSONSchemaDialect
	doc[openAPIPublisherExtension] = map[string]any{
		"format":          openAPIPublisherFormat,
		"artifact_type":   string(publicationKindModule),
		"module":          module.Name(),
		"normalized_from": openAPIBaselineVersion,
	}
	tagName := moduleTagName(comments, doc, module)
	tagDescription := moduleTagDescription(comments, doc, module)
	doc["tags"] = []any{map[string]any{
		"name":        tagName,
		"description": tagDescription,
	}}

	paths := ensureMap(doc, "paths")
	pathNames := sortedMapKeys(paths)
	for _, pathName := range pathNames {
		pathItem, ok := paths[pathName].(map[string]any)
		if !ok {
			continue
		}
		for _, method := range openAPIMethods() {
			operation, ok := pathItem[method].(map[string]any)
			if !ok {
				continue
			}
			operation["tags"] = []any{tagName}
		}
	}
}

// moduleTagName 返回模块文档最终采用的 tag 名称。
func moduleTagName(comments *protoCommentIndex, doc map[string]any, module moduleSpec) string {
	if comments != nil {
		name := strings.TrimSpace(comments.moduleTag.Name)
		if name != "" {
			return name
		}
	}
	tagName := firstExistingTagName(doc)
	if tagName == "" {
		tagName = module.Name()
	}
	return tagName
}

// moduleTagDescription 返回模块文档最终采用的 tag 说明。
func moduleTagDescription(comments *protoCommentIndex, doc map[string]any, module moduleSpec) string {
	if comments != nil {
		description := strings.TrimSpace(comments.moduleTag.Description)
		if description != "" {
			return description
		}
	}
	tagDescription := firstExistingTagDescription(doc, module)
	if tagDescription == "" {
		tagDescription = module.Description()
	}
	return tagDescription
}

// buildBundleDocument 聚合所有模块文档并输出最终 bundle。
func buildBundleDocument(modules []publishedDocument) (map[string]any, []byte, error) {
	bundle := map[string]any{
		"openapi":           openAPIReleaseVersion,
		"jsonSchemaDialect": openAPIJSONSchemaDialect,
		"info": map[string]any{
			"title":       "API Bundle",
			"version":     "v1",
			"description": "全量模块 OpenAPI 聚合文档。",
		},
		"paths":      map[string]any{},
		"components": map[string]any{},
		"tags":       []any{},
		openAPIPublisherExtension: map[string]any{
			"format":          openAPIPublisherFormat,
			"artifact_type":   string(publicationKindBundle),
			"bundle":          "openapi.json",
			"normalized_from": openAPIBaselineVersion,
		},
	}

	bundlePaths := ensureMap(bundle, "paths")
	bundleComponents := ensureMap(bundle, "components")
	tagsByName := map[string]map[string]any{}

	for _, module := range modules {
		docPaths := ensureMap(module.Document, "paths")
		for _, pathName := range sortedMapKeys(docPaths) {
			pathItem, ok := docPaths[pathName].(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("path item %s in %s is not an object", pathName, module.Module.Name())
			}
			if existing, exists := bundlePaths[pathName]; exists {
				merged, err := mergePathItem(existing.(map[string]any), pathItem, module.Module.Name(), pathName)
				if err != nil {
					return nil, nil, err
				}
				bundlePaths[pathName] = merged
				continue
			}
			bundlePaths[pathName] = cloneValue(pathItem)
		}

		docComponents, ok := module.Document["components"].(map[string]any)
		if ok {
			if err := mergeComponentGroups(bundleComponents, docComponents, module.Module.Name()); err != nil {
				return nil, nil, err
			}
		}

		for _, tag := range extractTags(module.Document) {
			name := strings.TrimSpace(asString(tag["name"]))
			if name == "" {
				continue
			}
			existing, exists := tagsByName[name]
			if !exists {
				tagsByName[name] = copyMap(tag)
				continue
			}
			same, err := sameJSON(existing, tag)
			if err != nil {
				return nil, nil, err
			}
			if !same {
				return nil, nil, fmt.Errorf("bundle tag conflict for %s", name)
			}
		}
	}

	tagNames := make([]string, 0, len(tagsByName))
	for name := range tagsByName {
		tagNames = append(tagNames, name)
	}
	sort.Strings(tagNames)
	tags := make([]any, 0, len(tagNames))
	for _, name := range tagNames {
		tags = append(tags, tagsByName[name])
	}
	bundle["tags"] = tags

	output, err := marshalCanonicalJSON(bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal bundle: %w", err)
	}
	return bundle, output, nil
}

// mergePathItem 合并同一路径下来自不同模块的 Path Item。
func mergePathItem(base map[string]any, patch map[string]any, moduleName, pathName string) (map[string]any, error) {
	result := copyMap(base)
	for key, value := range patch {
		existing, exists := result[key]
		if !exists {
			result[key] = cloneValue(value)
			continue
		}
		same, err := sameJSON(existing, value)
		if err != nil {
			return nil, err
		}
		if same {
			continue
		}
		if isOpenAPIMethod(key) {
			return nil, fmt.Errorf("bundle path conflict for %s %s in %s", strings.ToUpper(key), pathName, moduleName)
		}
		return nil, fmt.Errorf("bundle path metadata conflict for %s (%s) in %s", pathName, key, moduleName)
	}
	return result, nil
}

// mergeComponentGroups 合并 components 下各个命名子集合。
func mergeComponentGroups(bundleComponents map[string]any, docComponents map[string]any, moduleName string) error {
	for _, groupName := range sortedMapKeys(docComponents) {
		groupValue, ok := docComponents[groupName].(map[string]any)
		if !ok {
			continue
		}
		targetGroup := ensureMap(bundleComponents, groupName)
		for _, itemName := range sortedMapKeys(groupValue) {
			itemValue := groupValue[itemName]
			existing, exists := targetGroup[itemName]
			if !exists {
				targetGroup[itemName] = cloneValue(itemValue)
				continue
			}
			same, err := sameJSON(existing, itemValue)
			if err != nil {
				return err
			}
			if !same {
				return fmt.Errorf("bundle component conflict for %s.%s from %s", groupName, itemName, moduleName)
			}
		}
	}
	return nil
}

// extractTags 读取文档根级 tags。
func extractTags(doc map[string]any) []map[string]any {
	rawTags, ok := doc["tags"].([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(rawTags))
	for _, item := range rawTags {
		tag, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, copyMap(tag))
	}
	return result
}

// sameJSON 比较两个动态对象的 canonical JSON 是否一致。
func sameJSON(left, right any) (bool, error) {
	leftBytes, err := marshalCanonicalJSON(left)
	if err != nil {
		return false, err
	}
	rightBytes, err := marshalCanonicalJSON(right)
	if err != nil {
		return false, err
	}
	return string(leftBytes) == string(rightBytes), nil
}

// ensureMap 确保父对象中的指定键为 map，如不存在则创建。
func ensureMap(parent map[string]any, key string) map[string]any {
	if existing, ok := parent[key].(map[string]any); ok {
		return existing
	}
	created := map[string]any{}
	parent[key] = created
	return created
}

// setDefaultString 仅在字段缺失或为空时设置默认值。
func setDefaultString(target map[string]any, key, value string) {
	if strings.TrimSpace(asString(target[key])) != "" {
		return
	}
	target[key] = value
}

// sortedMapKeys 返回 map key 的稳定排序结果。
func sortedMapKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// openAPIMethods 返回 OpenAPI 支持的方法名集合。
func openAPIMethods() []string {
	return []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}
}

// isOpenAPIMethod 判断键名是否为 HTTP 方法。
func isOpenAPIMethod(key string) bool {
	for _, method := range openAPIMethods() {
		if key == method {
			return true
		}
	}
	return false
}
