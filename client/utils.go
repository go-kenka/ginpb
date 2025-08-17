package client

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// EncodeURL 将结构体字段编码到URL路径和查询参数中
// 类似Kratos的binding.EncodeURL功能
func EncodeURL(pathTemplate string, input interface{}, query bool) string {
	if input == nil {
		return pathTemplate
	}

	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return pathTemplate
	}

	t := v.Type()
	path := pathTemplate
	var queryParams []string

	// 遍历结构体字段
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 跳过无法访问的字段
		if !field.CanInterface() {
			continue
		}

		// 获取字段标签
		pathTag := fieldType.Tag.Get("uri")
		queryTag := fieldType.Tag.Get("form")

		// 处理路径参数
		if pathTag != "" {
			fieldValue := getFieldStringValue(field)
			if fieldValue != "" {
				placeholder := fmt.Sprintf("{%s}", pathTag)
				path = strings.ReplaceAll(path, placeholder, fieldValue)
			}
		}

		// 处理查询参数
		if query && queryTag != "" {
			fieldValue := getFieldStringValue(field)
			if fieldValue != "" {
				queryParams = append(queryParams, fmt.Sprintf("%s=%s",
					url.QueryEscape(queryTag), url.QueryEscape(fieldValue)))
			}
		}
	}

	// 添加查询参数
	if len(queryParams) > 0 {
		separator := "?"
		if strings.Contains(path, "?") {
			separator = "&"
		}
		path += separator + strings.Join(queryParams, "&")
	}

	return path
}

// getFieldStringValue 获取字段的字符串值
func getFieldStringValue(field reflect.Value) string {
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := field.Int()
		if val == 0 {
			return ""
		}
		return strconv.FormatInt(val, 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val := field.Uint()
		if val == 0 {
			return ""
		}
		return strconv.FormatUint(val, 10)
	case reflect.Float32, reflect.Float64:
		val := field.Float()
		if val == 0 {
			return ""
		}
		return strconv.FormatFloat(val, 'g', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(field.Bool())
	default:
		// 对于其他类型，尝试转换为字符串
		if field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
	}
	return ""
}

// BasicAuthValue 创建基础认证头值
func BasicAuthValue(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// JoinURL 连接URL路径
func JoinURL(base, path string) string {
	base = strings.TrimRight(base, "/")
	path = strings.TrimLeft(path, "/")
	if path == "" {
		return base
	}
	return base + "/" + path
}

// ParseEndpoint 解析端点URL
func ParseEndpoint(endpoint string) (*url.URL, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	// 如果没有协议，默认添加http://
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}

	return url.Parse(endpoint)
}

// BuildFullURL 构建完整的URL
func BuildFullURL(endpoint, path string) (string, error) {
	endpointURL, err := ParseEndpoint(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint: %w", err)
	}

	return JoinURL(endpointURL.String(), path), nil
}

// ExtractPathParams 从路径模板中提取参数名
func ExtractPathParams(pathTemplate string) []string {
	var params []string

	// 简单的正则替代方案
	parts := strings.Split(pathTemplate, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			param := strings.Trim(part, "{}")
			if param != "" {
				params = append(params, param)
			}
		}
	}

	return params
}

// ReplacePathParams 替换路径参数
func ReplacePathParams(pathTemplate string, params map[string]string) string {
	result := pathTemplate
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, url.PathEscape(value))
	}
	return result
}

// IsValidHTTPMethod 检查是否为有效的HTTP方法
func IsValidHTTPMethod(method string) bool {
	validMethods := []string{
		"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE",
	}

	method = strings.ToUpper(method)
	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}

// ContentTypeJSON JSON内容类型
const ContentTypeJSON = "application/json"

// ContentTypeForm 表单内容类型
const ContentTypeForm = "application/x-www-form-urlencoded"

// ContentTypeMultipart 多部分表单内容类型
const ContentTypeMultipart = "multipart/form-data"

// ContentTypeXML XML内容类型
const ContentTypeXML = "application/xml"

// ContentTypeText 文本内容类型
const ContentTypeText = "text/plain"
