package metadata

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ginKey struct{}

type GinData struct {
	Request *http.Request
	Params  gin.Params
	Writer  http.ResponseWriter
}

// NewContext put gin data into context
func NewContext(ctx *gin.Context) context.Context {
	data := &GinData{
		Request: ctx.Request,
		Params:  ctx.Params,
		Writer:  ctx.Writer,
	}
	return context.WithValue(ctx, ginKey{}, data)
}

// FromContext extract gin data from context
func FromContext(ctx context.Context) (data *GinData, ok bool) {
	data, ok = ctx.Value(ginKey{}).(*GinData)
	return
}
