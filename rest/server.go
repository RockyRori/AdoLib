package rest

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

// golangci-lint 要求独立定义key的类型
type key string

const XLangKey key = "X-Language"

const (
	XLangHeader     = "X-Language"
	ContentTypeKey  = "Content-Type"
	ContentTypeJson = "application/json"
)

// ReplyOK 响应成功。
func ReplyOK(c *gin.Context, statusCode int, body interface{}) {
	var bodyStr string
	if body != nil {
		b, _ := json.Marshal(body)
		bodyStr = string(b)
	}
	c.Writer.Header().Set(ContentTypeKey, ContentTypeJson)
	c.String(statusCode, bodyStr)
}

func ReplyOkWithHeaders(c *gin.Context, statusCode int, body interface{}, headers map[string]string) {
	addHeaders(c, headers)
	ReplyOK(c, statusCode, body)
}

// ReplyError 响应错误。
func ReplyError(c *gin.Context, err error) {
	var statusCode int
	var body string
	switch e := err.(type) {
	case *HTTPError:
		statusCode = e.HTTPCode
		body = e.Error()
	default:
		statusCode = http.StatusInternalServerError
		ctx := GetLanguageCtx(c)
		body = NewHTTPError(ctx, statusCode, InternalError).WithErrorDetails(e.Error()).Error()
	}

	c.Writer.Header().Set(ContentTypeKey, ContentTypeJson)
	c.String(statusCode, body)
}

func ReplyErrorWithHeaders(c *gin.Context, err error, headers map[string]string) {
	addHeaders(c, headers)
	ReplyError(c, err)
}

func addHeaders(c *gin.Context, headers map[string]string) {
	if len(headers) > 0 {
		for k, v := range headers {
			c.Writer.Header().Set(k, v)
		}
	}
}

func GetLanguageCtx(c *gin.Context) context.Context {
	langStr := c.GetHeader(XLangHeader)
	if langStr == "" {
		return context.WithValue(c.Request.Context(), XLangKey, "")
	}

	tags, _, err := language.ParseAcceptLanguage(langStr)
	if tags == nil || len(tags) != 1 || err != nil {
		log.Printf("invalid lang: %s", langStr)
		return context.WithValue(c.Request.Context(), XLangKey, "")
	}

	return context.WithValue(c.Request.Context(), XLangKey, tags[0].String())
}

func GetLanguageByCtx(ctx context.Context) string {
	lang := DefaultLanguage
	language := ctx.Value(XLangKey)
	if language != nil {
		lang = language.(string)
	}
	if _, ok := Languages[lang]; !ok {
		lang = DefaultLanguage
	}
	return lang
}
