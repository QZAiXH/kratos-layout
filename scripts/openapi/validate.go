package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	openapiv3 "github.com/google/gnostic/openapiv3"
	"go.yaml.in/yaml/v3"
)

// validatePublishedDocument 校验发布产物是否满足 3.1 发布契约。
func validatePublishedDocument(doc map[string]any, target publicationTarget) error {
	if strings.TrimSpace(asString(doc["openapi"])) != openAPIReleaseVersion {
		return fmt.Errorf("%s openapi version must be %s", target.Name, openAPIReleaseVersion)
	}
	if strings.TrimSpace(asString(doc["jsonSchemaDialect"])) != openAPIJSONSchemaDialect {
		return fmt.Errorf("%s jsonSchemaDialect must be %s", target.Name, openAPIJSONSchemaDialect)
	}
	publisher, ok := doc[openAPIPublisherExtension].(map[string]any)
	if !ok {
		return fmt.Errorf("%s missing %s metadata", target.Name, openAPIPublisherExtension)
	}
	if strings.TrimSpace(asString(publisher["format"])) != openAPIPublisherFormat {
		return fmt.Errorf("%s publisher format is invalid", target.Name)
	}
	if strings.TrimSpace(asString(publisher["artifact_type"])) != string(target.Kind) {
		return fmt.Errorf("%s publisher artifact_type must be %s", target.Name, target.Kind)
	}
	if strings.TrimSpace(asString(publisher["normalized_from"])) != openAPIBaselineVersion {
		return fmt.Errorf("%s normalized_from must be %s", target.Name, openAPIBaselineVersion)
	}

	parserInput := copyMap(doc)
	// gnostic v0.7.1 的 ParseDocument 仍无法识别 OpenAPI 3.1 根级 jsonSchemaDialect，
	// 因此这里在保留最终发布字段的前提下，使用去除此单个 3.1 字段的副本做结构解析。
	delete(parserInput, "jsonSchemaDialect")

	content, err := marshalCanonicalJSON(parserInput)
	if err != nil {
		return fmt.Errorf("marshal document for validation: %w", err)
	}
	if _, err := openapiv3.ParseDocument(content); err != nil {
		return fmt.Errorf("gnostic parse failed: %w", err)
	}
	if err := validateReferences(doc); err != nil {
		return err
	}
	if target.Kind == publicationKindModule {
		if err := validateSSEDocument(doc, target.Name); err != nil {
			return err
		}
	}
	return nil
}

// validateReferences 检查所有内部 $ref 都能在文档内被解析。
func validateReferences(doc map[string]any) error {
	return walkMaps(doc, func(_ []string, current any) error {
		mapped, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		ref := strings.TrimSpace(asString(mapped["$ref"]))
		if ref == "" {
			return nil
		}
		if !strings.HasPrefix(ref, "#/") {
			return fmt.Errorf("external ref is not allowed: %s", ref)
		}
		if _, ok := lookupJSONPointer(doc, ref); !ok {
			return fmt.Errorf("dangling ref: %s", ref)
		}
		return nil
	})
}

// validateSSEDocument 校验所有声明为 text/event-stream 的 operation 发布契约。
func validateSSEDocument(doc map[string]any, moduleName string) error {
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		return nil
	}
	for _, pathName := range sortedMapKeys(paths) {
		pathItem, ok := paths[pathName].(map[string]any)
		if !ok {
			continue
		}
		for _, method := range openAPIMethods() {
			operation, ok := pathItem[method].(map[string]any)
			if !ok || !operationHasSSEContent(operation) {
				continue
			}
			if err := validateSSEResponse(operation, []string{"Cache-Control", "Connection", "X-Accel-Buffering"}); err != nil {
				return fmt.Errorf("%s %s %s: %w", moduleName, strings.ToUpper(method), pathName, err)
			}
			if len(collectMediaExamples(operation)) == 0 {
				return fmt.Errorf("%s %s %s missing SSE example", moduleName, strings.ToUpper(method), pathName)
			}
		}
	}
	return nil
}

// operationHasSSEContent 判断 operation 是否声明了 text/event-stream 响应。
func operationHasSSEContent(operation map[string]any) bool {
	response, err := lookupResponse(operation, "200")
	if err != nil {
		return false
	}
	content, ok := response["content"].(map[string]any)
	if !ok {
		return false
	}
	_, ok = content["text/event-stream"].(map[string]any)
	return ok
}

// validateSSEResponse 校验 SSE 响应的 media type、schema 与关键 header。
func validateSSEResponse(operation map[string]any, requiredHeaders []string) error {
	response, err := lookupResponse(operation, "200")
	if err != nil {
		return err
	}
	content := ensureMap(response, "content")
	mediaType, ok := content["text/event-stream"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing text/event-stream response content")
	}
	schema, ok := mediaType["schema"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing text/event-stream schema")
	}
	if strings.TrimSpace(asString(schema["type"])) != "string" {
		return fmt.Errorf("text/event-stream schema.type must be string")
	}
	headers, ok := response["headers"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing SSE response headers")
	}
	for _, header := range requiredHeaders {
		if _, exists := headers[header]; !exists {
			return fmt.Errorf("missing SSE response header %s", header)
		}
	}
	return nil
}

// lookupResponse 读取 operation 下的指定响应。
func lookupResponse(operation map[string]any, code string) (map[string]any, error) {
	responses, ok := operation["responses"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("operation missing responses")
	}
	response, ok := responses[code].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("operation missing %s response", code)
	}
	return response, nil
}

// collectMediaExamples 汇总 text/event-stream 下的 example/examples 字段内容。
func collectMediaExamples(operation map[string]any) []string {
	response, err := lookupResponse(operation, "200")
	if err != nil {
		return nil
	}
	content, ok := response["content"].(map[string]any)
	if !ok {
		return nil
	}
	mediaType, ok := content["text/event-stream"].(map[string]any)
	if !ok {
		return nil
	}
	collected := make([]string, 0, 4)
	if example := strings.TrimSpace(asString(mediaType["example"])); example != "" {
		collected = append(collected, example)
	}
	if rawExamples, ok := mediaType["examples"].(map[string]any); ok {
		for _, name := range sortedMapKeys(rawExamples) {
			example, ok := rawExamples[name].(map[string]any)
			if !ok {
				continue
			}
			if value := strings.TrimSpace(asString(example["value"])); value != "" {
				collected = append(collected, value)
			}
		}
	}
	return collected
}

// lookupJSONPointer 根据 JSON Pointer 读取文档节点。
func lookupJSONPointer(doc map[string]any, ref string) (any, bool) {
	current := any(doc)
	for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch typed := current.(type) {
		case map[string]any:
			next, ok := typed[part]
			if !ok {
				return nil, false
			}
			current = next
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(typed) {
				return nil, false
			}
			current = typed[idx]
		default:
			return nil, false
		}
	}
	return current, true
}

// asString 把动态值转成字符串。
func asString(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(value)
	}
}

// unmarshalOpenAPI 兼容 YAML/JSON 输入的解码入口。
func unmarshalOpenAPI(content []byte, out any) error {
	if err := yaml.Unmarshal(content, out); err != nil {
		return fmt.Errorf("unmarshal openapi: %w", err)
	}
	return nil
}
