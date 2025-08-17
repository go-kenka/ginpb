package binding

import (
	"strings"

	"github.com/gin-gonic/gin"
	ginbinding "github.com/gin-gonic/gin/binding"
)

// BindByContentType automatically selects the appropriate binding method based on Content-Type header.
// This function supports all gin framework binding types including:
//   - JSON (default)
//   - XML
//   - YAML
//   - TOML
//   - ProtoBuf
//   - MsgPack
//   - Multipart Form
//   - URL-encoded Form
func BindByContentType(ctx *gin.Context, obj any) error {
	contentType := ctx.GetHeader("Content-Type")
	switch {
	case strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml"):
		return ctx.BindXML(obj)
	case strings.Contains(contentType, "application/x-yaml") || strings.Contains(contentType, "text/yaml"):
		return ctx.BindYAML(obj)
	case strings.Contains(contentType, "application/toml"):
		return ctx.BindTOML(obj)
	case strings.Contains(contentType, "application/x-protobuf"):
		return ctx.ShouldBindWith(obj, ginbinding.ProtoBuf)
	case strings.Contains(contentType, "application/x-msgpack"):
		return ctx.ShouldBindWith(obj, ginbinding.MsgPack)
	case strings.Contains(contentType, "multipart/form-data"):
		return ctx.ShouldBindWith(obj, ginbinding.FormMultipart)
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return ctx.Bind(obj)
	default:
		// Default to JSON binding for application/json and other content types
		return ctx.BindJSON(obj)
	}
}
