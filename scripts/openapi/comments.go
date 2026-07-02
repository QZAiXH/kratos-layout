package main

import (
	"fmt"
	"os"
	"strings"

	openapi_v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// protoCommentIndex 描述单个模块中可回填到 OpenAPI 的 proto 注释索引。
type protoCommentIndex struct {
	operations map[string]protoOperationComment // 按 method + path 建立的 operation 注释索引

	schemas map[string]protoSchemaComment // 按 OpenAPI schema 名建立的 schema 注释索引

	enums map[string]protoEnumComment // 按 OpenAPI schema 名建立的 enum 注释索引

	moduleTag protoTagComment // 模块级 tag 元信息
}

// protoOperationComment 描述单个 RPC 可回填的注释内容。
type protoOperationComment struct {
	Summary string // operation summary，默认取注释首行

	Description string // operation description，保留完整注释

	Parameters map[string]string // 参数名 -> 参数说明

	RequiredParameters map[string]struct{} // 参数名 -> 必填标记

	ParameterEnums map[string]protoEnumComment // 参数名 -> 枚举说明
}

// protoSchemaComment 描述单个 schema 可回填的注释内容。
type protoSchemaComment struct {
	Description string // schema 说明

	Properties map[string]string // 属性名 -> 属性说明

	PropertyEnums map[string]protoEnumComment // 属性名 -> 枚举说明
}

// protoEnumComment 描述单个 enum 可回填的注释内容。
type protoEnumComment struct {
	Description string // 枚举整体说明

	Values map[string]string // 枚举值名称 -> 枚举值说明
}

// enumValueDescription 描述单个枚举值可展示的说明。
type enumValueDescription struct {
	Name string // 枚举值名称

	Description string // 枚举值说明
}

// protoTagComment 描述模块级 tag 可回填的名称与说明。
type protoTagComment struct {
	Name string // tag 名称

	Description string // tag 说明
}

// buildProtoCommentIndex 从 descriptor set 中提取模块级 proto 注释索引。
func buildProtoCommentIndex(module moduleSpec, descriptorSetContent []byte) (*protoCommentIndex, error) {
	if len(descriptorSetContent) == 0 {
		return &protoCommentIndex{
			operations: map[string]protoOperationComment{},
			schemas:    map[string]protoSchemaComment{},
			enums:      map[string]protoEnumComment{},
			moduleTag:  protoTagComment{},
		}, nil
	}

	var descriptorSet descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(descriptorSetContent, &descriptorSet); err != nil {
		return nil, fmt.Errorf("unmarshal descriptor set: %w", err)
	}

	files, err := protodesc.NewFiles(&descriptorSet)
	if err != nil {
		return nil, fmt.Errorf("build file descriptors: %w", err)
	}

	index := &protoCommentIndex{
		operations: map[string]protoOperationComment{},
		schemas:    map[string]protoSchemaComment{},
		enums:      map[string]protoEnumComment{},
		moduleTag:  protoTagComment{},
	}
	for _, path := range module.ProtoFiles {
		file, err := findFileDescriptor(files, path)
		if err != nil {
			return nil, fmt.Errorf("find descriptor for %s: %w", path, err)
		}
		index.addFile(file)
	}

	return index, nil
}

// findFileDescriptor 兼容 protoc 仓库相对路径与 buf module 相对路径。
func findFileDescriptor(files *protoregistry.Files, path string) (protoreflect.FileDescriptor, error) {
	file, err := files.FindFileByPath(path)
	if err == nil {
		return file, nil
	}
	if strings.HasPrefix(path, "api/") {
		if file, trimmedErr := files.FindFileByPath(strings.TrimPrefix(path, "api/")); trimmedErr == nil {
			return file, nil
		}
	}
	return nil, err
}

// loadProtoCommentIndex 从 descriptor set 文件读取并构建注释索引。
func loadProtoCommentIndex(module moduleSpec, descriptorSetPath string) (*protoCommentIndex, error) {
	content, err := os.ReadFile(descriptorSetPath)
	if err != nil {
		return nil, fmt.Errorf("read descriptor set %s: %w", descriptorSetPath, err)
	}
	return buildProtoCommentIndex(module, content)
}

// addFile 提取单个 proto 文件中的可见注释。
func (i *protoCommentIndex) addFile(file protoreflect.FileDescriptor) {
	i.mergeModuleTag(fileDocumentTag(file))

	services := file.Services()
	for idx := 0; idx < services.Len(); idx++ {
		i.addService(services.Get(idx))
	}

	messages := file.Messages()
	for idx := 0; idx < messages.Len(); idx++ {
		i.addMessage(messages.Get(idx))
	}

	enums := file.Enums()
	for idx := 0; idx < enums.Len(); idx++ {
		i.addEnum(enums.Get(idx))
	}
}

// mergeModuleTag 合并模块级 tag 元信息，优先保留已存在的非空字段。
func (i *protoCommentIndex) mergeModuleTag(tag protoTagComment) {
	if strings.TrimSpace(i.moduleTag.Name) == "" {
		i.moduleTag.Name = strings.TrimSpace(tag.Name)
	}
	if strings.TrimSpace(i.moduleTag.Description) == "" {
		i.moduleTag.Description = strings.TrimSpace(tag.Description)
	}
}

// addService 提取 service 下所有 RPC 的注释映射。
func (i *protoCommentIndex) addService(service protoreflect.ServiceDescriptor) {
	methods := service.Methods()
	for idx := 0; idx < methods.Len(); idx++ {
		method := methods.Get(idx)
		rules := httpRulesForMethod(method)
		if len(rules) == 0 {
			continue
		}

		comment := descriptorComment(method)
		operationComment := protoOperationComment{
			Summary:            commentFirstLine(comment),
			Description:        comment,
			Parameters:         collectParameterComments(method.Input()),
			RequiredParameters: collectRequiredParameters(method.Input()),
			ParameterEnums:     collectParameterEnums(method.Input()),
		}

		for _, rule := range rules {
			methodName, pathName, ok := httpRuleOperation(rule)
			if !ok {
				continue
			}
			i.operations[openAPIOperationKey(methodName, pathName)] = operationComment
		}
	}
}

// addMessage 递归提取 message 及其字段注释。
func (i *protoCommentIndex) addMessage(message protoreflect.MessageDescriptor) {
	if message.IsMapEntry() {
		return
	}

	comment := descriptorComment(message)
	schema := protoSchemaComment{
		Description:   comment,
		Properties:    map[string]string{},
		PropertyEnums: map[string]protoEnumComment{},
	}

	fields := message.Fields()
	for idx := 0; idx < fields.Len(); idx++ {
		field := fields.Get(idx)
		fieldComment := fieldDescriptorComment(field)
		if fieldComment != "" {
			schema.Properties[string(field.Name())] = fieldComment
		}
		if field.Kind() == protoreflect.EnumKind {
			i.addEnum(field.Enum())
			enumComment := buildEnumComment(field.Enum())
			if enumComment.hasContent() {
				schema.PropertyEnums[string(field.Name())] = enumComment
			}
		}
	}
	i.schemas[openAPISchemaName(message)] = schema

	enums := message.Enums()
	for idx := 0; idx < enums.Len(); idx++ {
		i.addEnum(enums.Get(idx))
	}

	nestedMessages := message.Messages()
	for idx := 0; idx < nestedMessages.Len(); idx++ {
		i.addMessage(nestedMessages.Get(idx))
	}
}

// addEnum 提取 enum 及其枚举值注释。
func (i *protoCommentIndex) addEnum(enum protoreflect.EnumDescriptor) {
	comment := buildEnumComment(enum)
	if !comment.hasContent() {
		return
	}
	i.enums[openAPIEnumName(enum)] = comment
}

// httpRulesForMethod 返回 method 上声明的 HTTP 规则集合。
func httpRulesForMethod(method protoreflect.MethodDescriptor) []*annotations.HttpRule {
	options, ok := method.Options().(*descriptorpb.MethodOptions)
	if !ok || options == nil {
		return nil
	}
	extension := proto.GetExtension(options, annotations.E_Http)
	if extension == nil {
		return nil
	}

	rule, ok := extension.(*annotations.HttpRule)
	if !ok || rule == nil {
		return nil
	}

	return flattenHTTPRules(rule)
}

// flattenHTTPRules 展平主规则及 additional bindings。
func flattenHTTPRules(rule *annotations.HttpRule) []*annotations.HttpRule {
	if rule == nil {
		return nil
	}
	rules := []*annotations.HttpRule{rule}
	for _, item := range rule.AdditionalBindings {
		rules = append(rules, flattenHTTPRules(item)...)
	}
	return rules
}

// httpRuleOperation 把 HttpRule 转换成 OpenAPI method/path 键。
func httpRuleOperation(rule *annotations.HttpRule) (string, string, bool) {
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		return "get", pattern.Get, true
	case *annotations.HttpRule_Put:
		return "put", pattern.Put, true
	case *annotations.HttpRule_Post:
		return "post", pattern.Post, true
	case *annotations.HttpRule_Delete:
		return "delete", pattern.Delete, true
	case *annotations.HttpRule_Patch:
		return "patch", pattern.Patch, true
	default:
		return "", "", false
	}
}

// fileDocumentTag 返回 file options 中声明的模块级 tag 元信息。
func fileDocumentTag(file protoreflect.FileDescriptor) protoTagComment {
	options, ok := file.Options().(*descriptorpb.FileOptions)
	if !ok || options == nil || !proto.HasExtension(options, openapi_v3.E_Document) {
		return protoTagComment{}
	}
	extension := proto.GetExtension(options, openapi_v3.E_Document)
	document, ok := extension.(*openapi_v3.Document)
	if !ok || document == nil {
		return protoTagComment{}
	}
	for _, tag := range document.GetTags() {
		if tag == nil {
			continue
		}
		name := strings.TrimSpace(tag.GetName())
		description := strings.TrimSpace(tag.GetDescription())
		if name == "" && description == "" {
			continue
		}
		return protoTagComment{
			Name:        name,
			Description: description,
		}
	}
	return protoTagComment{}
}

// collectParameterComments 按 gnostic 的 query/path 命名规则收集参数说明。
func collectParameterComments(message protoreflect.MessageDescriptor) map[string]string {
	result := map[string]string{}
	fields := message.Fields()
	for idx := 0; idx < fields.Len(); idx++ {
		collectFieldParameterComments(result, fields.Get(idx), "")
	}
	return result
}

// collectRequiredParameters 按 gnostic 的 query/path 命名规则收集必填参数。
func collectRequiredParameters(message protoreflect.MessageDescriptor) map[string]struct{} {
	result := map[string]struct{}{}
	fields := message.Fields()
	for idx := 0; idx < fields.Len(); idx++ {
		collectFieldRequiredParameters(result, fields.Get(idx), "")
	}
	return result
}

// collectParameterEnums 按 gnostic 的 query/path 命名规则收集参数枚举说明。
func collectParameterEnums(message protoreflect.MessageDescriptor) map[string]protoEnumComment {
	result := map[string]protoEnumComment{}
	fields := message.Fields()
	for idx := 0; idx < fields.Len(); idx++ {
		collectFieldParameterEnums(result, fields.Get(idx), "")
	}
	return result
}

// collectFieldParameterComments 递归收集单个字段可映射出的参数说明。
func collectFieldParameterComments(target map[string]string, field protoreflect.FieldDescriptor, prefix string) {
	fieldName := string(field.Name())
	if prefix != "" {
		fieldName = prefix + "." + fieldName
	}
	fieldComment := fieldDescriptorComment(field)

	switch {
	case field.IsMap():
		return
	case field.Kind() == protoreflect.MessageKind:
		if shouldInlineMessageParameter(field) {
			setIfNotEmpty(target, fieldName, fieldComment)
			return
		}
		if field.IsList() {
			return
		}
		nestedFields := field.Message().Fields()
		for idx := 0; idx < nestedFields.Len(); idx++ {
			collectFieldParameterComments(target, nestedFields.Get(idx), fieldName)
		}
	default:
		setIfNotEmpty(target, fieldName, fieldComment)
	}
}

// collectFieldRequiredParameters 递归收集单个字段可映射出的必填参数。
func collectFieldRequiredParameters(target map[string]struct{}, field protoreflect.FieldDescriptor, prefix string) {
	fieldName := string(field.Name())
	if prefix != "" {
		fieldName = prefix + "." + fieldName
	}

	switch {
	case field.IsMap():
		return
	case field.Kind() == protoreflect.MessageKind:
		if shouldInlineMessageParameter(field) {
			setRequiredParameterIfNeeded(target, fieldName, field)
			return
		}
		if field.IsList() {
			return
		}
		nestedFields := field.Message().Fields()
		for idx := 0; idx < nestedFields.Len(); idx++ {
			collectFieldRequiredParameters(target, nestedFields.Get(idx), fieldName)
		}
	default:
		setRequiredParameterIfNeeded(target, fieldName, field)
	}
}

// collectFieldParameterEnums 递归收集单个字段可映射出的参数枚举说明。
func collectFieldParameterEnums(target map[string]protoEnumComment, field protoreflect.FieldDescriptor, prefix string) {
	fieldName := string(field.Name())
	if prefix != "" {
		fieldName = prefix + "." + fieldName
	}

	switch {
	case field.IsMap():
		return
	case field.Kind() == protoreflect.EnumKind:
		enumComment := buildEnumComment(field.Enum())
		if enumComment.hasContent() {
			target[fieldName] = enumComment
		}
	case field.Kind() == protoreflect.MessageKind:
		if shouldInlineMessageParameter(field) || field.IsList() {
			return
		}
		nestedFields := field.Message().Fields()
		for idx := 0; idx < nestedFields.Len(); idx++ {
			collectFieldParameterEnums(target, nestedFields.Get(idx), fieldName)
		}
	}
}

// shouldInlineMessageParameter 判断 message 字段是否应直接映射为单个参数。
func shouldInlineMessageParameter(field protoreflect.FieldDescriptor) bool {
	if field.Kind() != protoreflect.MessageKind {
		return false
	}

	typeName := string(field.Message().FullName())
	switch typeName {
	case "google.protobuf.Value",
		"google.protobuf.BoolValue",
		"google.protobuf.BytesValue",
		"google.protobuf.Int32Value",
		"google.protobuf.UInt32Value",
		"google.protobuf.StringValue",
		"google.protobuf.Int64Value",
		"google.protobuf.UInt64Value",
		"google.protobuf.FloatValue",
		"google.protobuf.DoubleValue",
		"google.protobuf.Timestamp",
		"google.protobuf.Duration",
		"google.protobuf.FieldMask":
		return true
	default:
		return false
	}
}

// enrichDocumentWithProtoComments 把 proto 注释回填到 OpenAPI 文档缺失位置。
func enrichDocumentWithProtoComments(doc map[string]any, comments *protoCommentIndex) {
	if comments == nil {
		return
	}

	enrichOperationComments(doc, comments.operations)
	enrichSchemaComments(doc, comments.schemas, comments.enums)
}

// enrichOperationComments 回填 operation 级别说明。
func enrichOperationComments(doc map[string]any, comments map[string]protoOperationComment) {
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		return
	}

	for _, pathName := range sortedMapKeys(paths) {
		pathItem, ok := paths[pathName].(map[string]any)
		if !ok {
			continue
		}
		for _, method := range openAPIMethods() {
			operation, ok := pathItem[method].(map[string]any)
			if !ok {
				continue
			}

			comment, ok := comments[openAPIOperationKey(method, pathName)]
			if !ok {
				continue
			}

			setDefaultString(operation, "summary", comment.Summary)
			setDefaultString(operation, "description", comment.Description)
			enrichParameterComments(operation, comment.Parameters, comment.RequiredParameters, comment.ParameterEnums)
		}
	}
}

// enrichParameterComments 回填 operation parameters 的说明与必填标记。
func enrichParameterComments(
	operation map[string]any,
	comments map[string]string,
	requiredParameters map[string]struct{},
	enumComments map[string]protoEnumComment,
) {
	rawParameters, ok := operation["parameters"].([]any)
	if !ok {
		return
	}

	for _, item := range rawParameters {
		parameter, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(asString(parameter["name"]))
		if name == "" {
			continue
		}
		setDefaultString(parameter, "description", comments[name])
		if _, ok := requiredParameters[name]; ok {
			parameter["required"] = true
		}
		enumComment, ok := enumComments[name]
		if !ok {
			continue
		}
		schema, ok := parameter["schema"].(map[string]any)
		if !ok {
			continue
		}
		enrichEnumValueComments(schema, enumComment)
		appendEnumValueDescriptionsFromSchema(parameter, schema, enumComment)
	}
}

// enrichSchemaComments 回填 components.schemas 及 properties 的说明。
func enrichSchemaComments(doc map[string]any, comments map[string]protoSchemaComment, enumComments map[string]protoEnumComment) {
	components, ok := doc["components"].(map[string]any)
	if !ok {
		return
	}
	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		return
	}

	for _, schemaName := range sortedMapKeys(schemas) {
		schema, ok := schemas[schemaName].(map[string]any)
		if !ok {
			continue
		}

		if enumComment, ok := enumComments[schemaName]; ok {
			enrichEnumValueComments(schema, enumComment)
		}

		comment, ok := comments[schemaName]
		if !ok {
			continue
		}
		setSchemaDescriptionIfMissing(schema, comment.Description)

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			continue
		}
		for _, propertyName := range sortedMapKeys(properties) {
			property, ok := properties[propertyName].(map[string]any)
			if !ok {
				continue
			}
			setSchemaDescriptionIfMissing(property, comment.Properties[propertyName])
			if enumComment, ok := comment.PropertyEnums[propertyName]; ok {
				enrichEnumValueComments(property, enumComment)
			}
		}
	}
}

// enrichEnumValueComments 回填 enum value 级别的扩展说明。
func enrichEnumValueComments(schema map[string]any, comment protoEnumComment) {
	if !comment.hasContent() {
		return
	}
	if items, ok := schema["items"].(map[string]any); ok {
		enrichEnumValueComments(items, comment)
		appendEnumValueDescriptionsFromSchema(schema, items, comment)
		return
	}

	rawValues, ok := schema["enum"].([]any)
	if !ok || len(rawValues) == 0 {
		setSchemaDescriptionIfMissing(schema, comment.Description)
		return
	}

	setSchemaDescriptionIfMissing(schema, comment.Description)

	varnames := make([]any, 0, len(rawValues))
	descriptions := make([]any, 0, len(rawValues))
	valueDescriptions := make([]enumValueDescription, 0, len(rawValues))
	for _, rawValue := range rawValues {
		name := strings.TrimSpace(asString(rawValue))
		description := strings.TrimSpace(comment.Values[name])
		varnames = append(varnames, name)
		descriptions = append(descriptions, description)
		if description == "" {
			continue
		}
		valueDescriptions = append(valueDescriptions, enumValueDescription{
			Name:        name,
			Description: description,
		})
	}
	if len(valueDescriptions) == 0 {
		return
	}
	setDefaultArray(schema, "x-enum-varnames", varnames)
	setDefaultArray(schema, "x-enum-descriptions", descriptions)
	appendEnumValueDescriptions(schema, valueDescriptions)
}

// appendEnumValueDescriptionsFromSchema 从 enum schema 中提取并追加枚举值说明。
func appendEnumValueDescriptionsFromSchema(target map[string]any, schema map[string]any, comment protoEnumComment) {
	if target == nil || schema == nil || !comment.hasContent() {
		return
	}
	if items, ok := schema["items"].(map[string]any); ok {
		appendEnumValueDescriptionsFromSchema(target, items, comment)
		return
	}
	rawValues, ok := schema["enum"].([]any)
	if !ok || len(rawValues) == 0 {
		return
	}
	valueDescriptions := make([]enumValueDescription, 0, len(rawValues))
	for _, rawValue := range rawValues {
		name := strings.TrimSpace(asString(rawValue))
		description := strings.TrimSpace(comment.Values[name])
		if description == "" {
			continue
		}
		valueDescriptions = append(valueDescriptions, enumValueDescription{
			Name:        name,
			Description: description,
		})
	}
	appendEnumValueDescriptions(target, valueDescriptions)
}

// appendEnumValueDescriptions 把枚举值说明追加到标准 description 字段。
func appendEnumValueDescriptions(target map[string]any, valueDescriptions []enumValueDescription) {
	if len(valueDescriptions) == 0 {
		return
	}
	block := renderEnumValueDescriptions(valueDescriptions)
	description := strings.TrimSpace(asString(target["description"]))
	if description == "" {
		target["description"] = block
		return
	}
	if strings.Contains(description, block) {
		return
	}
	target["description"] = description + "\n\n" + block
}

// renderEnumValueDescriptions 渲染面向 OpenAPI 查看器的枚举值说明块。
func renderEnumValueDescriptions(valueDescriptions []enumValueDescription) string {
	var builder strings.Builder
	builder.WriteString("枚举值：")
	for _, item := range valueDescriptions {
		builder.WriteString("\n- `")
		builder.WriteString(item.Name)
		builder.WriteString("`: ")
		builder.WriteString(item.Description)
	}
	return builder.String()
}

// setSchemaDescriptionIfMissing 仅在 description 缺失时补充 schema 说明。
func setSchemaDescriptionIfMissing(schema map[string]any, description string) {
	if strings.TrimSpace(description) == "" {
		return
	}
	if strings.TrimSpace(asString(schema["description"])) != "" {
		return
	}
	if ref := strings.TrimSpace(asString(schema["$ref"])); ref != "" {
		delete(schema, "$ref")
		schema["allOf"] = []any{map[string]any{"$ref": ref}}
	}
	schema["description"] = description
}

// openAPIOperationKey 返回 operation 的稳定索引键。
func openAPIOperationKey(method, path string) string {
	return strings.ToLower(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

// openAPISchemaName 按 gnostic 的 fq_schema_naming 规则生成 schema 名。
func openAPISchemaName(message protoreflect.MessageDescriptor) string {
	return string(message.ParentFile().Package()) + "." + openAPIMessageName(message)
}

// openAPIEnumName 按 gnostic 的 fq_schema_naming 规则生成 enum schema 名。
func openAPIEnumName(enum protoreflect.EnumDescriptor) string {
	return string(enum.ParentFile().Package()) + "." + openAPIEnumSimpleName(enum)
}

// openAPIEnumSimpleName 按 gnostic 的命名规则生成 enum 名。
func openAPIEnumSimpleName(enum protoreflect.EnumDescriptor) string {
	name := string(enum.Name())
	parent, ok := enum.Parent().(protoreflect.MessageDescriptor)
	if ok {
		name = openAPIMessageName(parent) + "_" + name
	}
	return name
}

// openAPIMessageName 按 gnostic 的命名规则生成 message 名。
func openAPIMessageName(message protoreflect.MessageDescriptor) string {
	name := string(message.Name())
	parent, ok := message.Parent().(protoreflect.MessageDescriptor)
	if ok {
		name = string(parent.Name()) + "_" + name
	}
	return name
}

// descriptorComment 返回描述符的首选注释文本。
func descriptorComment(descriptor protoreflect.Descriptor) string {
	location := descriptor.ParentFile().SourceLocations().ByDescriptor(descriptor)
	if text := leadingComment(location); text != "" {
		return text
	}
	return cleanCommentText(location.TrailingComments)
}

// fieldDescriptorComment 返回字段注释，优先 leading，其次 trailing。
func fieldDescriptorComment(field protoreflect.FieldDescriptor) string {
	location := field.ParentFile().SourceLocations().ByDescriptor(field)
	if text := leadingComment(location); text != "" {
		return text
	}
	return cleanCommentText(location.TrailingComments)
}

// leadingComment 拼接 detached + leading 注释块。
func leadingComment(location protoreflect.SourceLocation) string {
	parts := make([]string, 0, len(location.LeadingDetachedComments)+1)
	for _, item := range location.LeadingDetachedComments {
		if text := cleanCommentText(item); text != "" {
			parts = append(parts, text)
		}
	}
	if text := cleanCommentText(location.LeadingComments); text != "" {
		parts = append(parts, text)
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

// cleanCommentText 清理 source info 中的注释文本。
func cleanCommentText(value string) string {
	return strings.TrimSpace(value)
}

// commentFirstLine 返回注释中的首个非空行。
func commentFirstLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

// buildEnumComment 构造 enum 与 enum value 注释。
func buildEnumComment(enum protoreflect.EnumDescriptor) protoEnumComment {
	comment := protoEnumComment{
		Description: descriptorComment(enum),
		Values:      map[string]string{},
	}
	values := enum.Values()
	for idx := 0; idx < values.Len(); idx++ {
		value := values.Get(idx)
		valueComment := descriptorComment(value)
		if valueComment == "" {
			continue
		}
		comment.Values[string(value.Name())] = valueComment
	}
	return comment
}

// hasContent 判断 enum 注释是否包含可发布内容。
func (c protoEnumComment) hasContent() bool {
	if strings.TrimSpace(c.Description) != "" {
		return true
	}
	for _, value := range c.Values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

// firstExistingTagName 返回文档中已有 tag 的首个有效名称。
func firstExistingTagName(doc map[string]any) string {
	for _, tag := range extractTags(doc) {
		name := strings.TrimSpace(asString(tag["name"]))
		if name != "" {
			return name
		}
	}
	return ""
}

// firstExistingTagDescription 返回文档中已有 tag 的优先说明。
func firstExistingTagDescription(doc map[string]any, module moduleSpec) string {
	fallback := ""
	for _, tag := range extractTags(doc) {
		description := strings.TrimSpace(asString(tag["description"]))
		if description == "" {
			continue
		}
		if strings.TrimSpace(asString(tag["name"])) == module.Name() {
			return description
		}
		if fallback == "" {
			fallback = description
		}
	}
	return fallback
}

// setIfNotEmpty 仅在值非空时写入 map。
func setIfNotEmpty(target map[string]string, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	target[key] = value
}

// setDefaultArray 仅在字段缺失或为空数组时设置默认数组值。
func setDefaultArray(target map[string]any, key string, value []any) {
	if existing, ok := target[key].([]any); ok && len(existing) > 0 {
		return
	}
	target[key] = value
}

// setRequiredParameterIfNeeded 仅在字段显式声明 REQUIRED 时写入参数索引。
func setRequiredParameterIfNeeded(target map[string]struct{}, key string, field protoreflect.FieldDescriptor) {
	if !fieldHasBehavior(field, annotations.FieldBehavior_REQUIRED) {
		return
	}
	target[key] = struct{}{}
}

// fieldHasBehavior 判断字段是否声明了指定的 field_behavior。
func fieldHasBehavior(field protoreflect.FieldDescriptor, behavior annotations.FieldBehavior) bool {
	options, ok := field.Options().(*descriptorpb.FieldOptions)
	if !ok || options == nil || !proto.HasExtension(options, annotations.E_FieldBehavior) {
		return false
	}
	extension := proto.GetExtension(options, annotations.E_FieldBehavior)
	behaviors, ok := extension.([]annotations.FieldBehavior)
	if !ok {
		return false
	}
	for _, item := range behaviors {
		if item == behavior {
			return true
		}
	}
	return false
}
