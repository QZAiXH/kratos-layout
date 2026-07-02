package main

import (
	"testing"

	openapi_v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildProtoCommentIndexExtractsRPCAndFieldComments 验证 descriptor source info 会被提取为 operation/schema 注释索引。
func TestBuildProtoCommentIndexExtractsRPCAndFieldComments(t *testing.T) {
	descriptorSet := buildTestDescriptorSet(t)
	module := moduleSpec{
		Domain:     "demo",
		API:        "admin",
		Version:    "v1",
		ProtoFiles: []string{"api/demo/v1/demo.proto"},
	}

	index, err := buildProtoCommentIndex(module, descriptorSet)
	require.NoError(t, err)

	operation := index.operations["get /v1/things/{thing_id}"]
	assert.Equal(t, "获取详情。", operation.Summary)
	assert.Equal(t, "获取详情。\n第二行说明。", operation.Description)
	assert.Equal(t, "资源 ID", operation.Parameters["thing_id"])
	assert.Equal(t, "关键字", operation.Parameters["filter.keyword"])
	assert.Equal(t, "状态", operation.Parameters["status"])
	_, requiredThingID := operation.RequiredParameters["thing_id"]
	assert.True(t, requiredThingID)
	_, requiredKeyword := operation.RequiredParameters["filter.keyword"]
	assert.True(t, requiredKeyword)
	assert.Equal(t, "状态枚举。", operation.ParameterEnums["status"].Description)
	assert.Equal(t, "启用", operation.ParameterEnums["status"].Values["THING_STATUS_ACTIVE"])

	requestSchema := index.schemas["demo.v1.GetThingRequest"]
	assert.Equal(t, "获取详情请求。", requestSchema.Description)
	assert.Equal(t, "资源 ID", requestSchema.Properties["thing_id"])
	assert.Equal(t, "筛选条件。", requestSchema.Properties["filter"])
	assert.Equal(t, "状态", requestSchema.Properties["status"])
	assert.Equal(t, "禁用", requestSchema.PropertyEnums["status"].Values["THING_STATUS_DISABLED"])

	filterSchema := index.schemas["demo.v1.Filter"]
	assert.Equal(t, "筛选条件实体。", filterSchema.Description)
	assert.Equal(t, "关键字", filterSchema.Properties["keyword"])

	replySchema := index.schemas["demo.v1.GetThingReply"]
	assert.Equal(t, "获取详情响应。", replySchema.Description)
	assert.Equal(t, "名称", replySchema.Properties["name"])
	assert.Equal(t, "状态", replySchema.Properties["status"])
	assert.Equal(t, "外部状态", replySchema.Properties["external_state"])
	assert.Equal(t, "状态列表", replySchema.Properties["states"])
	assert.Equal(t, "就绪", replySchema.PropertyEnums["external_state"].Values["EXTERNAL_STATE_READY"])
	assert.Equal(t, "状态枚举。", index.enums["demo.v1.ThingStatus"].Description)
	_, hasUnspecifiedComment := index.enums["demo.v1.ThingStatus"].Values["THING_STATUS_UNSPECIFIED"]
	assert.False(t, hasUnspecifiedComment)
	assert.Equal(t, "外部状态枚举。", index.enums["demo.common.v1.ExternalState"].Description)
	assert.Equal(t, "演示模块", index.moduleTag.Name)
	assert.Equal(t, "演示模块接口。", index.moduleTag.Description)
}

// TestEnrichDocumentWithProtoCommentsPreservesExistingDescriptions 验证 proto 注释仅补空位，不覆盖显式 OpenAPI 注解。
func TestEnrichDocumentWithProtoCommentsPreservesExistingDescriptions(t *testing.T) {
	descriptorSet := buildTestDescriptorSet(t)
	module := moduleSpec{
		Domain:     "demo",
		API:        "admin",
		Version:    "v1",
		ProtoFiles: []string{"api/demo/v1/demo.proto"},
	}

	index, err := buildProtoCommentIndex(module, descriptorSet)
	require.NoError(t, err)

	doc := map[string]any{
		"paths": map[string]any{
			"/v1/things/{thing_id}": map[string]any{
				"get": map[string]any{
					"summary":     "",
					"description": "",
					"parameters": []any{
						map[string]any{"name": "thing_id", "in": "path"},
						map[string]any{"name": "filter.keyword", "in": "query", "description": "显式参数说明", "required": false},
						map[string]any{"name": "status", "in": "query", "schema": map[string]any{
							"type":   "string",
							"format": "enum",
							"enum": []any{
								"THING_STATUS_UNSPECIFIED",
								"THING_STATUS_ACTIVE",
								"THING_STATUS_DISABLED",
							},
						}},
						map[string]any{"name": "page_size", "in": "query"},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "OK"},
					},
				},
			},
		},
		"components": map[string]any{
			"schemas": map[string]any{
				"demo.v1.GetThingRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"thing_id": map[string]any{"type": "string"},
						"filter":   map[string]any{"$ref": "#/components/schemas/demo.v1.Filter"},
						"status": map[string]any{
							"type":   "string",
							"format": "enum",
							"enum": []any{
								"THING_STATUS_UNSPECIFIED",
								"THING_STATUS_ACTIVE",
								"THING_STATUS_DISABLED",
							},
						},
					},
				},
				"demo.v1.Filter": map[string]any{
					"type":        "object",
					"description": "显式 schema 说明",
					"properties": map[string]any{
						"keyword": map[string]any{"type": "string", "description": "显式属性说明"},
					},
				},
				"demo.v1.GetThingReply": map[string]any{
					"type":        "object",
					"description": "",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"status": map[string]any{
							"type":   "string",
							"format": "enum",
							"enum": []any{
								"THING_STATUS_UNSPECIFIED",
								"THING_STATUS_ACTIVE",
								"THING_STATUS_DISABLED",
							},
						},
						"external_state": map[string]any{
							"type":   "string",
							"format": "enum",
							"enum": []any{
								"EXTERNAL_STATE_UNKNOWN",
								"EXTERNAL_STATE_READY",
							},
						},
						"states": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type":   "string",
								"format": "enum",
								"enum": []any{
									"THING_STATUS_UNSPECIFIED",
									"THING_STATUS_ACTIVE",
									"THING_STATUS_DISABLED",
								},
							},
						},
					},
				},
				"demo.v1.ThingStatus": map[string]any{
					"type":   "string",
					"format": "enum",
					"enum": []any{
						"THING_STATUS_UNSPECIFIED",
						"THING_STATUS_ACTIVE",
						"THING_STATUS_DISABLED",
					},
				},
			},
		},
	}

	enrichDocumentWithProtoComments(doc, index)

	operation := doc["paths"].(map[string]any)["/v1/things/{thing_id}"].(map[string]any)["get"].(map[string]any)
	assert.Equal(t, "获取详情。", operation["summary"])
	assert.Equal(t, "获取详情。\n第二行说明。", operation["description"])

	parameters := operation["parameters"].([]any)
	assert.Equal(t, "资源 ID", parameters[0].(map[string]any)["description"])
	assert.Equal(t, true, parameters[0].(map[string]any)["required"])
	assert.Equal(t, "显式参数说明", parameters[1].(map[string]any)["description"])
	assert.Equal(t, true, parameters[1].(map[string]any)["required"])
	statusParameterSchema := parameters[2].(map[string]any)["schema"].(map[string]any)
	assert.Equal(t, []any{"THING_STATUS_UNSPECIFIED", "THING_STATUS_ACTIVE", "THING_STATUS_DISABLED"}, statusParameterSchema["x-enum-varnames"])
	assert.Equal(t, []any{"", "启用", "禁用"}, statusParameterSchema["x-enum-descriptions"])
	thingStatusValueDescription := "枚举值：\n- `THING_STATUS_ACTIVE`: 启用\n- `THING_STATUS_DISABLED`: 禁用"
	assert.Equal(t, "状态\n\n"+thingStatusValueDescription, parameters[2].(map[string]any)["description"])
	assert.Equal(t, "状态枚举。\n\n"+thingStatusValueDescription, statusParameterSchema["description"])
	_, hasRequired := parameters[3].(map[string]any)["required"]
	assert.False(t, hasRequired)

	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	requestSchema := schemas["demo.v1.GetThingRequest"].(map[string]any)
	assert.Equal(t, "获取详情请求。", requestSchema["description"])
	assert.Equal(t, "资源 ID", requestSchema["properties"].(map[string]any)["thing_id"].(map[string]any)["description"])

	filterProperty := requestSchema["properties"].(map[string]any)["filter"].(map[string]any)
	assert.Equal(t, "筛选条件。", filterProperty["description"])
	assert.NotContains(t, filterProperty, "$ref")
	assert.Equal(t, []any{map[string]any{"$ref": "#/components/schemas/demo.v1.Filter"}}, filterProperty["allOf"])
	requestStatusProperty := requestSchema["properties"].(map[string]any)["status"].(map[string]any)
	assert.Equal(t, "状态\n\n"+thingStatusValueDescription, requestStatusProperty["description"])
	assert.Equal(t, []any{"", "启用", "禁用"}, requestStatusProperty["x-enum-descriptions"])

	filterSchema := schemas["demo.v1.Filter"].(map[string]any)
	assert.Equal(t, "显式 schema 说明", filterSchema["description"])
	assert.Equal(t, "显式属性说明", filterSchema["properties"].(map[string]any)["keyword"].(map[string]any)["description"])

	replySchema := schemas["demo.v1.GetThingReply"].(map[string]any)
	assert.Equal(t, "获取详情响应。", replySchema["description"])
	assert.Equal(t, "名称", replySchema["properties"].(map[string]any)["name"].(map[string]any)["description"])
	replyStatusProperty := replySchema["properties"].(map[string]any)["status"].(map[string]any)
	assert.Equal(t, "状态\n\n"+thingStatusValueDescription, replyStatusProperty["description"])
	replyExternalStateProperty := replySchema["properties"].(map[string]any)["external_state"].(map[string]any)
	assert.Equal(t, "外部状态\n\n枚举值：\n- `EXTERNAL_STATE_UNKNOWN`: 未知\n- `EXTERNAL_STATE_READY`: 就绪", replyExternalStateProperty["description"])
	assert.Equal(t, []any{"EXTERNAL_STATE_UNKNOWN", "EXTERNAL_STATE_READY"}, replyExternalStateProperty["x-enum-varnames"])
	assert.Equal(t, []any{"未知", "就绪"}, replyExternalStateProperty["x-enum-descriptions"])
	replyStatesProperty := replySchema["properties"].(map[string]any)["states"].(map[string]any)
	assert.Equal(t, "状态列表\n\n"+thingStatusValueDescription, replyStatesProperty["description"])
	replyStatesItems := replyStatesProperty["items"].(map[string]any)
	assert.Equal(t, "状态枚举。\n\n"+thingStatusValueDescription, replyStatesItems["description"])
	assert.Equal(t, []any{"", "启用", "禁用"}, replyStatesItems["x-enum-descriptions"])

	statusSchema := schemas["demo.v1.ThingStatus"].(map[string]any)
	assert.Equal(t, "状态枚举。\n\n"+thingStatusValueDescription, statusSchema["description"])
	assert.Equal(t, []any{"", "启用", "禁用"}, statusSchema["x-enum-descriptions"])
}

// TestPublishModuleDocumentPreservesExistingTagMetadata 验证模块规范化时保留 proto 中已有的 tag 名称与说明。
func TestPublishModuleDocumentPreservesExistingTagMetadata(t *testing.T) {
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
tags:
  - name: 演示服务
    description: 演示服务
`)
	module := moduleSpec{Domain: "demo", API: "admin", Version: "v1"}

	doc, _, err := publishModuleDocument(module, baseline, nil, nil)
	require.NoError(t, err)

	tags := doc["tags"].([]any)
	tag := tags[0].(map[string]any)
	assert.Equal(t, "演示服务", tag["name"])
	assert.Equal(t, "演示服务", tag["description"])

	operation := doc["paths"].(map[string]any)["/ping"].(map[string]any)["get"].(map[string]any)
	assert.Equal(t, []any{"演示服务"}, operation["tags"])
}

// TestPublishModuleDocumentUsesProtoDocumentTagMetadata 验证模块规范化优先采用 proto document 声明的 tag 元信息。
func TestPublishModuleDocumentUsesProtoDocumentTagMetadata(t *testing.T) {
	baseline := []byte(`
openapi: 3.0.3
info:
  title: Demo
  version: v1
paths:
  /ping:
    get:
      tags:
        - DemoService
      responses:
        "200":
          description: OK
tags:
  - name: DemoService
    description: 演示服务
`)
	module := moduleSpec{Domain: "demo", API: "admin", Version: "v1"}
	comments := &protoCommentIndex{
		moduleTag: protoTagComment{
			Name:        "演示模块",
			Description: "演示模块接口。",
		},
	}

	doc, _, err := publishModuleDocument(module, baseline, nil, comments)
	require.NoError(t, err)

	tags := doc["tags"].([]any)
	tag := tags[0].(map[string]any)
	assert.Equal(t, "演示模块", tag["name"])
	assert.Equal(t, "演示模块接口。", tag["description"])

	operation := doc["paths"].(map[string]any)["/ping"].(map[string]any)["get"].(map[string]any)
	assert.Equal(t, []any{"演示模块"}, operation["tags"])
}

// buildTestDescriptorSet 构造最小 descriptor set，供注释提取逻辑测试使用。
func buildTestDescriptorSet(t *testing.T) []byte {
	t.Helper()

	methodOptions := &descriptorpb.MethodOptions{}
	proto.SetExtension(methodOptions, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Get{Get: "/v1/things/{thing_id}"},
	})
	requiredFieldOptions := &descriptorpb.FieldOptions{}
	proto.SetExtension(requiredFieldOptions, annotations.E_FieldBehavior, []annotations.FieldBehavior{annotations.FieldBehavior_REQUIRED})
	fileOptions := &descriptorpb.FileOptions{}
	proto.SetExtension(fileOptions, openapi_v3.E_Document, &openapi_v3.Document{
		Tags: []*openapi_v3.Tag{{
			Name:        "演示模块",
			Description: "演示模块接口。",
		}},
	})

	externalFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("api/demo/v1/common.proto"),
		Package: proto.String("demo.common.v1"),
		Syntax:  proto.String("proto3"),
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: proto.String("ExternalState"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{
						Name:   proto.String("EXTERNAL_STATE_UNKNOWN"),
						Number: proto.Int32(0),
					},
					{
						Name:   proto.String("EXTERNAL_STATE_READY"),
						Number: proto.Int32(1),
					},
				},
			},
		},
		SourceCodeInfo: &descriptorpb.SourceCodeInfo{
			Location: []*descriptorpb.SourceCodeInfo_Location{
				{Path: []int32{5, 0}, Span: []int32{0, 0, 0}, LeadingComments: proto.String("外部状态枚举。\n")},
				{Path: []int32{5, 0, 2, 0}, Span: []int32{1, 0, 1}, TrailingComments: proto.String("未知\n")},
				{Path: []int32{5, 0, 2, 1}, Span: []int32{2, 0, 2}, TrailingComments: proto.String("就绪\n")},
			},
		},
	}

	file := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("api/demo/v1/demo.proto"),
		Package:    proto.String("demo.v1"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"google/api/annotations.proto", "google/api/field_behavior.proto", "openapiv3/annotations.proto", "api/demo/v1/common.proto"},
		Options:    fileOptions,
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("GetThingRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:    proto.String("thing_id"),
						Number:  proto.Int32(1),
						Label:   descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Options: requiredFieldOptions,
					},
					{
						Name:     proto.String("filter"),
						Number:   proto.Int32(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".demo.v1.Filter"),
					},
					{
						Name:     proto.String("status"),
						Number:   proto.Int32(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: proto.String(".demo.v1.ThingStatus"),
					},
				},
			},
			{
				Name: proto.String("Filter"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:    proto.String("keyword"),
						Number:  proto.Int32(1),
						Label:   descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Options: requiredFieldOptions,
					},
				},
			},
			{
				Name: proto.String("GetThingReply"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("name"),
						Number: proto.Int32(1),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:     proto.String("status"),
						Number:   proto.Int32(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: proto.String(".demo.v1.ThingStatus"),
					},
					{
						Name:     proto.String("external_state"),
						Number:   proto.Int32(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: proto.String(".demo.common.v1.ExternalState"),
					},
					{
						Name:     proto.String("states"),
						Number:   proto.Int32(4),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: proto.String(".demo.v1.ThingStatus"),
					},
				},
			},
		},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: proto.String("ThingStatus"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{
						Name:   proto.String("THING_STATUS_UNSPECIFIED"),
						Number: proto.Int32(0),
					},
					{
						Name:   proto.String("THING_STATUS_ACTIVE"),
						Number: proto.Int32(1),
					},
					{
						Name:   proto.String("THING_STATUS_DISABLED"),
						Number: proto.Int32(2),
					},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("DemoService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("GetThing"),
						InputType:  proto.String(".demo.v1.GetThingRequest"),
						OutputType: proto.String(".demo.v1.GetThingReply"),
						Options:    methodOptions,
					},
				},
			},
		},
		SourceCodeInfo: &descriptorpb.SourceCodeInfo{
			Location: []*descriptorpb.SourceCodeInfo_Location{
				{Path: []int32{4, 0}, Span: []int32{0, 0, 0}, LeadingComments: proto.String("获取详情请求。\n")},
				{Path: []int32{4, 0, 2, 0}, Span: []int32{1, 0, 1}, TrailingComments: proto.String("资源 ID\n")},
				{Path: []int32{4, 0, 2, 1}, Span: []int32{2, 0, 2}, LeadingComments: proto.String("筛选条件。\n")},
				{Path: []int32{4, 0, 2, 2}, Span: []int32{3, 0, 3}, TrailingComments: proto.String("状态\n")},
				{Path: []int32{4, 1}, Span: []int32{3, 0, 3}, LeadingComments: proto.String("筛选条件实体。\n")},
				{Path: []int32{4, 1, 2, 0}, Span: []int32{4, 0, 4}, TrailingComments: proto.String("关键字\n")},
				{Path: []int32{4, 2}, Span: []int32{5, 0, 5}, LeadingComments: proto.String("获取详情响应。\n")},
				{Path: []int32{4, 2, 2, 0}, Span: []int32{6, 0, 6}, TrailingComments: proto.String("名称\n")},
				{Path: []int32{4, 2, 2, 1}, Span: []int32{7, 0, 7}, TrailingComments: proto.String("状态\n")},
				{Path: []int32{4, 2, 2, 2}, Span: []int32{8, 0, 8}, TrailingComments: proto.String("外部状态\n")},
				{Path: []int32{4, 2, 2, 3}, Span: []int32{9, 0, 9}, TrailingComments: proto.String("状态列表\n")},
				{Path: []int32{5, 0}, Span: []int32{9, 0, 9}, LeadingComments: proto.String("状态枚举。\n")},
				{Path: []int32{5, 0, 2, 1}, Span: []int32{11, 0, 11}, TrailingComments: proto.String("启用\n")},
				{Path: []int32{5, 0, 2, 2}, Span: []int32{12, 0, 12}, TrailingComments: proto.String("禁用\n")},
				{Path: []int32{6, 0}, Span: []int32{7, 0, 7}, LeadingComments: proto.String("演示服务。\n")},
				{Path: []int32{6, 0, 2, 0}, Span: []int32{8, 0, 8}, LeadingComments: proto.String("获取详情。\n第二行说明。\n")},
			},
		},
	}

	set := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			protodesc.ToFileDescriptorProto(descriptorpb.File_google_protobuf_descriptor_proto),
			protodesc.ToFileDescriptorProto(anypb.File_google_protobuf_any_proto),
			protodesc.ToFileDescriptorProto(annotations.File_google_api_http_proto),
			protodesc.ToFileDescriptorProto(annotations.File_google_api_annotations_proto),
			protodesc.ToFileDescriptorProto(annotations.File_google_api_field_behavior_proto),
			protodesc.ToFileDescriptorProto(openapi_v3.File_openapiv3_OpenAPIv3_proto),
			protodesc.ToFileDescriptorProto(openapi_v3.File_openapiv3_annotations_proto),
			externalFile,
			file,
		},
	}

	content, err := proto.Marshal(set)
	require.NoError(t, err)
	return content
}
