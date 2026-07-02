package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiscoverModules 验证模块发现会排除 error proto，并按字典序输出稳定模块名。
func TestDiscoverModules(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "api/user/admin/v1"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "api/marketview/interface/v1"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "api/todo/v1"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "api/user/admin/v1", "user_admin.proto"), []byte("syntax = \"proto3\";"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "api/user/admin/v1", "user_admin_error.proto"), []byte("syntax = \"proto3\";"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "api/marketview/interface/v1", "marketview_interface.proto"), []byte("syntax = \"proto3\";"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "api/marketview/interface/v1", "marketview_interface_sse.proto"), []byte("syntax = \"proto3\";"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "api/todo/v1", "todo.proto"), []byte("syntax = \"proto3\";"), 0o644))

	modules, err := discoverModules(root)
	require.NoError(t, err)
	require.Len(t, modules, 3)
	assert.Equal(t, []string{"marketview_interface_v1", "todo_v1", "user_admin_v1"}, []string{modules[0].Name(), modules[1].Name(), modules[2].Name()})
	assert.Equal(t, []string{"api/marketview/interface/v1/marketview_interface.proto", "api/marketview/interface/v1/marketview_interface_sse.proto"}, modules[0].ProtoFiles)
	assert.Equal(t, []string{"api/todo/v1/todo.proto"}, modules[1].ProtoFiles)
	assert.Equal(t, []string{"api/user/admin/v1/user_admin.proto"}, modules[2].ProtoFiles)
}

// TestMergeDocumentSupportsDeletionAndReplacement 验证 overlay 可以删除 */* 并替换为 SSE content。
func TestMergeDocumentSupportsDeletionAndReplacement(t *testing.T) {
	base := map[string]any{
		"paths": map[string]any{
			"/stream": map[string]any{
				"get": map[string]any{
					"responses": map[string]any{
						"200": map[string]any{
							"content": map[string]any{
								"*/*": map[string]any{"schema": map[string]any{"type": "string"}},
							},
						},
					},
				},
			},
		},
	}
	overlay := map[string]any{
		"paths": map[string]any{
			"/stream": map[string]any{
				"get": map[string]any{
					"responses": map[string]any{
						"200": map[string]any{
							"content": map[string]any{
								"*/*":               nil,
								"text/event-stream": map[string]any{"schema": map[string]any{"type": "string"}},
							},
						},
					},
				},
			},
		},
	}

	merged, err := mergeDocument(base, overlay)
	require.NoError(t, err)
	content := merged["paths"].(map[string]any)["/stream"].(map[string]any)["get"].(map[string]any)["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)
	assert.NotContains(t, content, "*/*")
	assert.Contains(t, content, "text/event-stream")
}

// TestValidatePublishedDocumentRejectsPseudo31 验证“只改版本号”的伪 3.1 文档会被发布校验拒绝。
func TestValidatePublishedDocumentRejectsPseudo31(t *testing.T) {
	doc := map[string]any{
		"openapi": "3.1.0",
		"info":    map[string]any{"title": "Demo", "version": "v1"},
		"paths": map[string]any{
			"/ping": map[string]any{
				"get": map[string]any{
					"responses": map[string]any{"200": map[string]any{"description": "OK"}},
				},
			},
		},
	}

	err := validatePublishedDocument(doc, publicationTarget{Name: "demo", Kind: publicationKindModule})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jsonSchemaDialect")
}

// TestPublishModuleDocumentProducesStableJSON 验证同一输入重复发布得到的 JSON 字节级一致。
func TestPublishModuleDocumentProducesStableJSON(t *testing.T) {
	baseline := []byte(`
openapi: 3.0.3
info:
  title: Demo
  version: v1
paths:
  /ping:
    get:
      responses:
        "200":
          description: OK
components:
  schemas:
    Demo:
      type: object
`)
	module := moduleSpec{Domain: "file", API: "interface", Version: "v1"}

	firstDoc, firstJSON, err := publishModuleDocument(module, baseline, nil, nil)
	require.NoError(t, err)
	secondDoc, secondJSON, err := publishModuleDocument(module, baseline, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, string(firstJSON), string(secondJSON))
	assert.Equal(t, openAPIReleaseVersion, firstDoc["openapi"])
	assert.Equal(t, openAPIBaselineVersion, firstDoc[openAPIPublisherExtension].(map[string]any)["normalized_from"])
	assert.Equal(t, firstDoc, secondDoc)
}

// TestBuildBundleDocumentDedupesSameSchemaAndRejectsConflicts 验证 bundle 会对同名同内容 schema 去重，并在冲突时失败。
func TestBuildBundleDocumentDedupesSameSchemaAndRejectsConflicts(t *testing.T) {
	first := publishedDocument{
		Module: moduleSpec{Domain: "user", API: "admin", Version: "v1"},
		Document: map[string]any{
			"openapi":           openAPIReleaseVersion,
			"jsonSchemaDialect": openAPIJSONSchemaDialect,
			"info":              map[string]any{"title": "first", "version": "v1"},
			"paths": map[string]any{
				"/first": map[string]any{"get": map[string]any{"responses": map[string]any{"200": map[string]any{"description": "OK"}}}},
			},
			"components": map[string]any{"schemas": map[string]any{"Shared": map[string]any{"type": "object"}}},
			"tags":       []any{map[string]any{"name": "user_admin_v1"}},
		},
	}
	second := publishedDocument{
		Module: moduleSpec{Domain: "user", API: "interface", Version: "v1"},
		Document: map[string]any{
			"openapi":           openAPIReleaseVersion,
			"jsonSchemaDialect": openAPIJSONSchemaDialect,
			"info":              map[string]any{"title": "second", "version": "v1"},
			"paths": map[string]any{
				"/second": map[string]any{"get": map[string]any{"responses": map[string]any{"200": map[string]any{"description": "OK"}}}},
			},
			"components": map[string]any{"schemas": map[string]any{"Shared": map[string]any{"type": "object"}}},
			"tags":       []any{map[string]any{"name": "user_interface_v1"}},
		},
	}

	bundle, _, err := buildBundleDocument([]publishedDocument{first, second})
	require.NoError(t, err)
	schemas := bundle["components"].(map[string]any)["schemas"].(map[string]any)
	schemaKeys := mustMapKeys(schemas)
	sort.Strings(schemaKeys)
	assert.Equal(t, []string{"Shared"}, schemaKeys)

	conflicting := publishedDocument{
		Module: moduleSpec{Domain: "marketview", API: "interface", Version: "v1"},
		Document: map[string]any{
			"openapi":           openAPIReleaseVersion,
			"jsonSchemaDialect": openAPIJSONSchemaDialect,
			"info":              map[string]any{"title": "third", "version": "v1"},
			"paths": map[string]any{
				"/third": map[string]any{"get": map[string]any{"responses": map[string]any{"200": map[string]any{"description": "OK"}}}},
			},
			"components": map[string]any{"schemas": map[string]any{"Shared": map[string]any{"type": "string"}}},
		},
	}

	_, _, err = buildBundleDocument([]publishedDocument{first, conflicting})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bundle component conflict")
}

// TestValidateReferencesRejectsDanglingRef 验证悬空引用会被校验器拒绝。
func TestValidateReferencesRejectsDanglingRef(t *testing.T) {
	doc := map[string]any{
		"openapi":           openAPIReleaseVersion,
		"jsonSchemaDialect": openAPIJSONSchemaDialect,
		"info":              map[string]any{"title": "Demo", "version": "v1"},
		openAPIPublisherExtension: map[string]any{
			"format":          openAPIPublisherFormat,
			"artifact_type":   string(publicationKindModule),
			"normalized_from": openAPIBaselineVersion,
		},
		"paths": map[string]any{
			"/ping": map[string]any{
				"get": map[string]any{
					"responses": map[string]any{
						"200": map[string]any{
							"description": "OK",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{"$ref": "#/components/schemas/Missing"},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]any{"schemas": map[string]any{}},
	}

	err := validatePublishedDocument(doc, publicationTarget{Name: "demo", Kind: publicationKindModule})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dangling ref")
}

// mustMapKeys 返回 map 的 key 列表，供测试断言使用。
func mustMapKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	return keys
}
