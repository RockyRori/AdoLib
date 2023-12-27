package rest

import (
	"context"
	"encoding/json"
	"log"

	. "github.com/RockyRori/AdoLib/i18n"
)

type BaseError struct {
	ErrorCode               string                 `json:"error_code"`    // 错误码
	Description             string                 `json:"description"`   // 错误描述
	Solution                string                 `json:"solution"`      // 解决方法
	ErrorLink               string                 `json:"error_link"`    // 错误链接
	ErrorDetails            interface{}            `json:"error_details"` // 详细内容
	DescriptionTemplateData map[string]interface{} `json:"-"`             // 错误描述参数
	SolutionTemplateData    map[string]interface{} `json:"-"`             // 解决方法参数
}

var (
	// Languages 支持的语言
	Languages = map[string]string{
		"zh-CN": "zh-CN",
		"en-US": "en-US",
	}
	DefaultLanguage = "zh-CN"

	allErrs = commonErrorI18n
)

// SetLang 设置语言
func SetLang(langStr string) {
	if _, ok := Languages[langStr]; !ok {
		log.Fatalf("invalid lang: %s", langStr)
	}

	DefaultLanguage = langStr
}

func Register(errorCodeList []string) {
	for _, errorCode := range errorCodeList {
		if _, ok := allErrs[errorCode]; ok {
			log.Fatalf("duplicate errorCode: %s", errorCode)
		}
		allErrs[errorCode] = make(map[string]BaseError)
		for lang := range Languages {
			allErrs[errorCode][lang] = BaseError{
				ErrorCode:               errorCode,
				Description:             Translate(lang, errorCode+".Description", nil),
				Solution:                Translate(lang, errorCode+".Solution", nil),
				ErrorLink:               Translate(lang, errorCode+".ErrorLink", nil),
				ErrorDetails:            "",
				DescriptionTemplateData: make(map[string]interface{}),
				SolutionTemplateData:    make(map[string]interface{}),
			}
		}
	}
}

type HTTPError struct {
	HTTPCode  int
	Language  string
	BaseError BaseError
}

// NewHTTPError 创建 HTTPError。
func NewHTTPError(ctx context.Context, httpCode int, errorCode string) *HTTPError {
	lang := GetLanguageByCtx(ctx)

	errs, ok := allErrs[errorCode]
	if !ok {
		log.Fatalf("missing errorCode: %s", errorCode)
		return nil
	}
	err := errs[lang]
	if !ok {
		log.Fatalf("errorCode %s missing lang: %s", errorCode, lang)
		return nil
	}

	return &HTTPError{
		HTTPCode: httpCode,
		Language: lang,
		BaseError: BaseError{
			ErrorCode:    errorCode,
			Description:  err.Description,
			ErrorLink:    err.ErrorLink,
			Solution:     err.Solution,
			ErrorDetails: err.ErrorDetails,
		},
	}
}

func (e *HTTPError) WithDescription(templateData map[string]interface{}) *HTTPError {
	e.BaseError.DescriptionTemplateData = templateData
	e.BaseError.Description = Translate(e.Language, e.BaseError.ErrorCode+".Description", templateData)
	return e
}

func (e *HTTPError) WithSolution(templateData map[string]interface{}) *HTTPError {
	e.BaseError.SolutionTemplateData = templateData
	e.BaseError.Solution = Translate(e.Language, e.BaseError.ErrorCode+".Solution", templateData)
	return e
}

// WithErrorDetails 设置错误详情。
func (e *HTTPError) WithErrorDetails(errorDetails interface{}) *HTTPError {
	e.BaseError.ErrorDetails = errorDetails
	return e
}

func (e *HTTPError) Error() string {
	errStr, _ := json.Marshal(e.BaseError)
	return string(errStr)
}
