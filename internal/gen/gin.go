package gen

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	ginext "github.com/go-kenka/ginpb/tag"
)

const (
	contextPackage     = protogen.GoImportPath("context")
	ginPackage         = protogen.GoImportPath("github.com/gin-gonic/gin")
	bindingPackage     = protogen.GoImportPath("github.com/gin-gonic/gin/binding")
	bindingutilPackage = protogen.GoImportPath("github.com/go-kenka/ginpb/binding")
	metadataPackage    = protogen.GoImportPath("github.com/go-kenka/ginpb/metadata")
	middlewarePackage  = protogen.GoImportPath("github.com/go-kenka/ginpb/middleware")
	clientPackage      = protogen.GoImportPath("github.com/go-kenka/ginpb/client")
	fmtPackage         = protogen.GoImportPath("fmt")
	stringsPackage     = protogen.GoImportPath("strings")
)

var serverTemplate = `{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

{{- range .MethodSets}}
const Operation{{$svrType}}{{.OriginalName}} = "/{{$svrName}}/{{.OriginalName}}"
{{- end}}

type {{.ServiceType}}HTTPServer interface {
{{- range .MethodSets}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}HTTPServer(r gin.IRouter, srv {{.ServiceType}}HTTPServer) {
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))
	{{- end}}
}

func Register{{.ServiceType}}HTTPServerWithMiddleware(r gin.IRouter, srv {{.ServiceType}}HTTPServer, middlewares ...gin.HandlerFunc) {
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", append(middlewares, _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))...)
	{{- end}}
}

func Register{{.ServiceType}}HTTPServerWithOperationMiddleware(r gin.IRouter, srv {{.ServiceType}}HTTPServer, middlewares map[string][]gin.HandlerFunc) {
	{{- range .Methods}}
	if mws, exists := middlewares[Operation{{$svrType}}{{.OriginalName}}]; exists {
		r.{{.Method}}("{{.Path}}", append(mws, _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))...)
	} else {
		r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))
	}
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{$svrType}}HTTPServer) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// Set operation for middleware
		ctx.Set("operation", Operation{{$svrType}}{{.OriginalName}})
		
		{{if .Fields}}var ginReq {{.Name | lower}}GinRequest{{else}}var in {{.Request}}{{end}}
		{{- if .HasBody}}
		// body binding with automatic Content-Type detection
		{{if .Fields}}if err := binding1.BindByContentType(ctx, &ginReq); err != nil {
		{{else}}if err := binding1.BindByContentType(ctx, &in); err != nil {
		{{- end}}
			ctx.Error(err)
			return
		}
		
		{{- if not (eq .Body "")}}
		// query
		{{if .Fields}}if err := ctx.BindQuery(&ginReq); err != nil {
		{{else}}if err := ctx.BindQuery(&in); err != nil {
		{{- end}}
			ctx.Error(err)
			return
		}
		{{- end}}
		{{- else}}
		// query
		{{if .Fields}}if err := ctx.BindQuery(&ginReq); err != nil {
		{{else}}if err := ctx.BindQuery(&in); err != nil {
		{{- end}}
			ctx.Error(err)
			return
		}
		{{- end}}
		{{- if .HasParams}}
		// params
		{{if .Fields}}if err := ctx.BindUri(&ginReq); err != nil {
		{{else}}if err := ctx.BindUri(&in); err != nil {
		{{- end}}
			ctx.Error(err)
			return
		}
		{{- end}}
		{{if .Fields}}
		// Convert gin request to protobuf request
		in := ginReq.to{{.Name}}Request()
		
		// Custom field tags detected:
		{{range .Fields}}
		// Field {{.GoName}}: {{range $key, $value := .Tags}}{{$key}}:"{{$value}}" {{end}}
		{{- end}}
		{{- end}}
		// header,ip等常用信息, form表单信息,包括上传文件
		newCtx := metadata.NewContext(ctx)
		{{if .Fields}}reply, err := srv.{{.Name}}(newCtx, in){{else}}reply, err := srv.{{.Name}}(newCtx, &in){{end}}
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(200, reply{{.ResponseBody}})
	}
}
{{end}}`

var clientTemplate = `{{$svrType := .ServiceType}}

type {{.ServiceType}}HTTPClient interface {
{{- range .MethodSets}}
	{{.Name}}(ctx context.Context, req *{{.Request}}, opts ...client.CallOption) (rsp *{{.Reply}}, err error) 
{{- end}}
}
	
type {{.ServiceType}}HTTPClientImpl struct{
	client client.Client
}
	
func New{{.ServiceType}}HTTPClient(opts ...client.ClientOption) {{.ServiceType}}HTTPClient {
	c := client.NewClient(opts...)
	return &{{.ServiceType}}HTTPClientImpl{client: c}
}

{{range .MethodSets}}
func (c *{{$svrType}}HTTPClientImpl) {{.Name}}(ctx context.Context, in *{{.Request}}, opts ...client.CallOption) (*{{.Reply}}, error) {
	var out {{.Reply}}
	
	// 构建请求路径
	path := "{{.ClientPath}}"
	{{- if .HasParams}}
	// 替换路径参数
	{{- range .PathParams}}
	path = strings.ReplaceAll(path, "{{print "{" . "}" }}", fmt.Sprintf("%v", in.{{camelCase .}}))
	{{- end}}
	{{- end}}
	
	{{- if eq .Method "GET"}}
	// GET请求
	err := c.client.Invoke(ctx, "{{.Method}}", path, nil, &out{{.ResponseBody}}, opts...)
	{{- else}}
	// {{.Method}}请求
	{{if .HasBody -}}
	err := c.client.Invoke(ctx, "{{.Method}}", path, in{{.Body}}, &out{{.ResponseBody}}, opts...)
	{{else -}} 
	err := c.client.Invoke(ctx, "{{.Method}}", path, nil, &out{{.ResponseBody}}, opts...)
	{{end -}}
	{{- end}}
	
	if err != nil {
		return nil, fmt.Errorf("{{.Method}} {{.ClientPath}} failed: %w", err)
	}
	return &out, nil
}
{{end}}`

var tagsStructTemplate = `// Internal structs with gin binding tags for protobuf messages
{{$svrType := .ServiceType}}
{{range .MethodSets}}
{{if .Fields}}
// {{.Name | lower}}GinRequest provides gin binding tags for {{.Request}}
type {{.Name | lower}}GinRequest struct {
{{range .Fields}}	{{.GoName}} {{.GoType}} {{formatTags .Tags}}
{{end}}}

// convert{{.Name}}GinRequest converts from gin request struct to protobuf struct
func (r *{{.Name | lower}}GinRequest) to{{.Name}}Request() *{{.Request}} {
	return &{{.Request}}{
{{range .Fields}}		{{.GoName}}: r.{{.GoName}},
{{end}}	}
}

// from{{.Name}}Request converts from protobuf struct to gin request struct  
func from{{.Name}}Request(req *{{.Request}}) *{{.Name | lower}}GinRequest {
	return &{{.Name | lower}}GinRequest{
{{range .Fields}}		{{.GoName}}: req.{{.GoName}},
{{end}}	}
}
{{end}}
{{end}}`

const Release = "v1.0.0" // Plugin version

var methodSets = make(map[string]int)

// GenerateFile generates a .pb.gin.go file using resty-based client
func GenerateFile(gen *protogen.Plugin, file *protogen.File, omitempty bool) *protogen.GeneratedFile {
	if len(file.Services) == 0 || (omitempty && !hasHTTPRule(file.Services)) {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + ".pb.gin.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-gin with resty client. DO NOT EDIT.")
	g.P("// versions:")
	g.P(fmt.Sprintf("// - protoc-gen-gin %s", Release))
	g.P("// - protoc             ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	generateFileContent(gen, file, g, omitempty)
	return g
}

// generateFileContent generates the resty-based client implementation
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, omitempty bool) {
	if len(file.Services) == 0 {
		return
	}
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the resty client it is being compiled against.")
	g.P("var _ = new(", contextPackage.Ident("Context"), ")")
	g.P("var _ = new(", metadataPackage.Ident("GinData"), ")")
	g.P("var _ = new(", ginPackage.Ident("H"), ")")
	g.P("var _ = new(", clientPackage.Ident("Client"), ")")
	g.P("var _ = ", bindingPackage.Ident("JSON"))
	g.P("var _ = ", bindingutilPackage.Ident("BindByContentType"))
	g.P("var _ = ", middlewarePackage.Ident("Chain"))
	g.P("var _ = ", fmtPackage.Ident("Sprintf"))
	g.P("var _ = ", stringsPackage.Ident("ReplaceAll"))
	g.P()

	for _, service := range file.Services {
		genService(gen, file, g, service, omitempty)
	}
}

func genService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, omitempty bool) {
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}

	// HTTP Server.
	sd := &serviceDesc{
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		Metadata:    file.Desc.Path(),
	}
	for _, method := range service.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}
		rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
		if rule != nil && ok {
			for _, bind := range rule.AdditionalBindings {
				sd.Methods = append(sd.Methods, buildHTTPRule(g, method, bind))
			}
			sd.Methods = append(sd.Methods, buildHTTPRule(g, method, rule))
		} else if !omitempty {
			path := fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())
			sd.Methods = append(sd.Methods, buildMethodDesc(g, method, http.MethodPost, path))
		}
	}
	if len(sd.Methods) != 0 {
		g.P(sd.execute())
	}
}

func buildHTTPRule(g *protogen.GeneratedFile, m *protogen.Method, rule *annotations.HttpRule) *methodDesc {
	var (
		path         string
		method       string
		body         string
		responseBody string
	)

	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = http.MethodGet
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = http.MethodPut
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = http.MethodPost
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = http.MethodDelete
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = http.MethodPatch
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	body = rule.Body
	responseBody = rule.ResponseBody
	md := buildMethodDesc(g, m, method, path)

	// 解析路径参数
	md.PathParams = extractPathParams(path)

	if method == http.MethodGet || method == http.MethodDelete {
		if body != "" {
			_, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s body should not be declared.\n", method, path)
		}
	} else {
		if body == "" {
			_, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s does not declare a body.\n", method, path)
		}
	}
	if body == "*" {
		md.HasBody = true
		md.Body = ""
	} else if body != "" {
		md.HasBody = true
		md.Body = "." + camelCaseVars(body)
	} else {
		md.HasBody = false
	}
	if responseBody == "*" {
		md.ResponseBody = ""
	} else if responseBody != "" {
		md.ResponseBody = "." + camelCaseVars(responseBody)
	}
	return md
}

func buildMethodDesc(g *protogen.GeneratedFile, m *protogen.Method, method, path string) *methodDesc {
	defer func() { methodSets[m.GoName]++ }()

	params := buildPathParams(path)

	for v, s := range params {
		fields := m.Input.Desc.Fields()

		if s != nil {
			path = replacePath(v, *s, path)
		}
		for _, field := range strings.Split(v, ".") {
			if strings.TrimSpace(field) == "" {
				continue
			}
			if strings.Contains(field, ":") {
				field = strings.Split(field, ":")[0]
			}
			fd := fields.ByName(protoreflect.Name(field))
			if fd == nil {
				fmt.Fprintf(os.Stderr, "\u001B[31mERROR\u001B[m: The corresponding field '%s' declaration in message could not be found in '%s'\n", v, path)
				os.Exit(2)
			}
			if fd.IsMap() {
				fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field in path:'%s' shouldn't be a map.\n", v)
			} else if fd.IsList() {
				fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field in path:'%s' shouldn't be a list.\n", v)
			} else if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
				fields = fd.Message().Fields()
			}
		}
	}
	return &methodDesc{
		Name:         m.GoName,
		OriginalName: string(m.Desc.Name()),
		Num:          methodSets[m.GoName],
		Request:      g.QualifiedGoIdent(m.Input.GoIdent),
		Reply:        g.QualifiedGoIdent(m.Output.GoIdent),
		Path:         transformPath(path),
		ClientPath:   path,
		Method:       method,
		HasParams:    len(params) > 0,
		Fields:       parseMessageFields(m.Input),
	}
}

// 辅助函数
func extractPathParams(path string) []string {
	pattern := regexp.MustCompile(`{([^}]+)}`)
	matches := pattern.FindAllStringSubmatch(path, -1)
	var params []string
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}
	return params
}

// parseFieldTags parses custom gin tags from field options
func parseFieldTags(field *protogen.Field) map[string]string {
	tags := make(map[string]string)

	opts := field.Desc.Options().(*descriptorpb.FieldOptions)

	// Parse structured tags
	if fieldTags, ok := proto.GetExtension(opts, ginext.E_Tags).(*ginext.FieldTags); ok && fieldTags != nil {
		if form := fieldTags.GetForm(); form != "" {
			tags["form"] = form
		}
		if uri := fieldTags.GetUri(); uri != "" {
			tags["uri"] = uri
		}
		if json := fieldTags.GetJson(); json != "" {
			tags["json"] = json
		}
		if header := fieldTags.GetHeader(); header != "" {
			tags["header"] = header
		}
		if binding := fieldTags.GetBinding(); binding != "" {
			tags["binding"] = binding
		}
		if validate := fieldTags.GetValidate(); validate != "" {
			tags["validate"] = validate
		}
		if xml := fieldTags.GetXml(); xml != "" {
			tags["xml"] = xml
		}
		if yaml := fieldTags.GetYaml(); yaml != "" {
			tags["yaml"] = yaml
		}
		if toml := fieldTags.GetToml(); toml != "" {
			tags["toml"] = toml
		}
		if protobuf := fieldTags.GetProtobuf(); protobuf != "" {
			tags["protobuf"] = protobuf
		}
		if msgpack := fieldTags.GetMsgpack(); msgpack != "" {
			tags["msgpack"] = msgpack
		}
		if multipart := fieldTags.GetMultipart(); multipart != "" {
			tags["multipart"] = multipart
		}
		if custom := fieldTags.GetCustom(); custom != "" {
			// Parse custom tags in format "key1:value1,key2:value2"
			pairs := strings.Split(custom, ",")
			for _, pair := range pairs {
				if kv := strings.SplitN(pair, ":", 2); len(kv) == 2 {
					tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}
			}
		}
	}

	// Parse shortcut tags
	if formTag, ok := proto.GetExtension(opts, ginext.E_FormTag).(string); ok && formTag != "" {
		tags["form"] = formTag
	}
	if uriTag, ok := proto.GetExtension(opts, ginext.E_UriTag).(string); ok && uriTag != "" {
		tags["uri"] = uriTag
	}
	if headerTag, ok := proto.GetExtension(opts, ginext.E_HeaderTag).(string); ok && headerTag != "" {
		tags["header"] = headerTag
	}
	if bindingTag, ok := proto.GetExtension(opts, ginext.E_BindingTag).(string); ok && bindingTag != "" {
		tags["binding"] = bindingTag
	}
	if xmlTag, ok := proto.GetExtension(opts, ginext.E_XmlTag).(string); ok && xmlTag != "" {
		tags["xml"] = xmlTag
	}
	if yamlTag, ok := proto.GetExtension(opts, ginext.E_YamlTag).(string); ok && yamlTag != "" {
		tags["yaml"] = yamlTag
	}
	if tomlTag, ok := proto.GetExtension(opts, ginext.E_TomlTag).(string); ok && tomlTag != "" {
		tags["toml"] = tomlTag
	}
	if protobufTag, ok := proto.GetExtension(opts, ginext.E_ProtobufTag).(string); ok && protobufTag != "" {
		tags["protobuf"] = protobufTag
	}
	if msgpackTag, ok := proto.GetExtension(opts, ginext.E_MsgpackTag).(string); ok && msgpackTag != "" {
		tags["msgpack"] = msgpackTag
	}
	if multipartTag, ok := proto.GetExtension(opts, ginext.E_MultipartTag).(string); ok && multipartTag != "" {
		tags["multipart"] = multipartTag
	}

	// Auto-generate json tag if not explicitly set
	if _, hasJson := tags["json"]; !hasJson {
		tags["json"] = string(field.Desc.Name())
	}

	return tags
}

// getGoType converts protobuf field type to Go type string
func getGoType(field *protogen.Field) string {
	// Handle repeated fields (arrays/slices)
	if field.Desc.IsList() {
		elementType := getScalarGoType(field)
		return "[]" + elementType
	}

	// Handle map fields
	if field.Desc.IsMap() {
		keyType := getMapKeyType(field.Desc.MapKey())
		valueType := getMapValueType(field.Desc.MapValue())
		return fmt.Sprintf("map[%s]%s", keyType, valueType)
	}

	return getScalarGoType(field)
}

// getScalarGoType gets the Go type for scalar protobuf types
func getScalarGoType(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.EnumKind:
		return "int32" // Enums are typically int32 in Go
	case protoreflect.MessageKind:
		// For message types, we'll use the full Go type name
		return "*" + field.Message.GoIdent.GoName
	default:
		return "interface{}" // fallback
	}
}

// getMapKeyType returns the Go type for map keys
func getMapKeyType(keyField protoreflect.FieldDescriptor) string {
	switch keyField.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	default:
		return "string" // fallback to string
	}
}

// getMapValueType returns the Go type for map values
func getMapValueType(valueField protoreflect.FieldDescriptor) string {
	switch valueField.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.EnumKind:
		return "int32"
	case protoreflect.MessageKind:
		// For message types in maps, we don't use pointers
		return string(valueField.Message().Name())
	default:
		return "interface{}" // fallback
	}
}

// parseMessageFields recursively parses message fields and extracts tag information
func parseMessageFields(message *protogen.Message) []*fieldInfo {
	var fields []*fieldInfo

	for _, field := range message.Fields {
		fieldInfo := &fieldInfo{
			Name:     string(field.Desc.Name()),
			GoName:   field.GoName,
			GoType:   getGoType(field),
			JsonName: field.Desc.JSONName(),
			Tags:     parseFieldTags(field),
		}
		fields = append(fields, fieldInfo)

		// TODO: Handle nested messages if needed
	}

	return fields
}

// formatStructTags formats tag map into Go struct tag string
func formatStructTags(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}

	var parts []string
	// Order tags consistently: json, xml, yaml, toml, form, uri, header, protobuf, msgpack, multipart, binding, validate, custom
	tagOrder := []string{"json", "xml", "yaml", "toml", "form", "uri", "header", "protobuf", "msgpack", "multipart", "binding", "validate"}

	for _, key := range tagOrder {
		if value, ok := tags[key]; ok {
			parts = append(parts, fmt.Sprintf(`%s:"%s"`, key, value))
		}
	}

	// Add any remaining custom tags
	for key, value := range tags {
		found := false
		for _, orderKey := range tagOrder {
			if key == orderKey {
				found = true
				break
			}
		}
		if !found {
			parts = append(parts, fmt.Sprintf(`%s:"%s"`, key, value))
		}
	}

	if len(parts) > 0 {
		return "`" + strings.Join(parts, " ") + "`"
	}
	return ""
}

// hasTag checks if a field has a specific tag
func hasTag(field *fieldInfo, tagName string) bool {
	_, exists := field.Tags[tagName]
	return exists
}

// getTag gets tag value from a field
func getTag(field *fieldInfo, tagName string) string {
	return field.Tags[tagName]
}

func hasHTTPRule(services []*protogen.Service) bool {
	for _, service := range services {
		for _, method := range service.Methods {
			if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
				continue
			}
			rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
			if rule != nil && ok {
				return true
			}
		}
	}
	return false
}

// transformPath 转换参数路由 {xx} --> :xx
func transformPath(path string) string {
	paths := strings.Split(path, "/")
	for i, p := range paths {
		if len(p) > 0 && (p[0] == '{' && p[len(p)-1] == '}' || p[0] == ':') {
			paths[i] = ":" + p[1:len(p)-1]
		}
	}
	return strings.Join(paths, "/")
}

func buildPathParams(path string) (res map[string]*string) {
	if strings.HasSuffix(path, "/") {
		fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: Path %s should not end with \"/\" \n", path)
	}
	pattern := regexp.MustCompile(`(?i){([a-z.0-9_\s]*)=?([^{}]*)}`)
	matches := pattern.FindAllStringSubmatch(path, -1)
	res = make(map[string]*string, len(matches))
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if len(name) > 1 && len(m[2]) > 0 {
			res[name] = &m[2]
		} else {
			res[name] = nil
		}
	}
	return
}

func replacePath(name string, value string, path string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?i){([\s]*%s[\s]*)=?([^{}]*)}`, name))
	idx := pattern.FindStringIndex(path)
	if len(idx) > 0 {
		path = fmt.Sprintf("%s%s:%s%s",
			path[:idx[0]], // The start of the match
			name,
			strings.ReplaceAll(value, "*", ".*"),
			path[idx[1]:],
		)
	}
	return path
}

func camelCaseVars(s string) string {
	subs := strings.Split(s, ".")
	vars := make([]string, 0, len(subs))
	for _, sub := range subs {
		vars = append(vars, camelCase(sub))
	}
	return strings.Join(vars, ".")
}

// camelCase returns the CamelCased name.
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
// There is a remote possibility of this rewrite causing a name collision,
// but it's so remote we're prepared to pretend it's nonexistent - since the
// C++ generator lowercase names, it's extremely unlikely to have two fields
// with different capitalization.
// In short, _my_field_name_2 becomes XMyFieldName_2.
func camelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

type serviceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodDesc
	MethodSets  map[string]*methodDesc
}

type fieldInfo struct {
	Name     string
	GoName   string
	GoType   string
	JsonName string
	Tags     map[string]string // tag name -> tag value
}

type methodDesc struct {
	// method
	Name         string
	OriginalName string // The parsed original name
	Num          int
	Request      string
	Reply        string
	// http_rule
	Path         string
	Method       string
	HasParams    bool
	HasBody      bool
	Body         string
	ResponseBody string
	ClientPath   string
	// resty specific
	PathParams []string
	// field information for tag generation
	Fields []*fieldInfo
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}

	buf := new(bytes.Buffer)

	// Generate server code
	serverTmpl, err := template.New("server").Funcs(template.FuncMap{
		"camelCase":  camelCase,
		"formatTags": formatStructTags,
		"hasTag":     hasTag,
		"getTag":     getTag,
		"lower":      strings.ToLower,
	}).Parse(strings.TrimSpace(serverTemplate))
	if err != nil {
		panic(err)
	}
	if err := serverTmpl.Execute(buf, s); err != nil {
		panic(err)
	}

	buf.WriteString("\n\n")

	// Generate client code
	clientTmpl, err := template.New("client").Funcs(template.FuncMap{
		"camelCase": camelCase,
	}).Parse(strings.TrimSpace(clientTemplate))
	if err != nil {
		panic(err)
	}
	if err := clientTmpl.Execute(buf, s); err != nil {
		panic(err)
	}

	buf.WriteString("\n\n")

	// Generate tagged structs at the end
	tagsTmpl, err := template.New("tags").Funcs(template.FuncMap{
		"formatTags": formatStructTags,
		"lower":      strings.ToLower,
	}).Parse(strings.TrimSpace(tagsStructTemplate))
	if err != nil {
		panic(err)
	}
	if err := tagsTmpl.Execute(buf, s); err != nil {
		panic(err)
	}

	return strings.Trim(buf.String(), "\r\n")
}

const deprecationComment = "// Deprecated: Do not use."
