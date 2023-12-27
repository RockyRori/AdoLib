package rest

// 系统默认错误
const (
	// InternalError 通用错误码，服务端内部错误
	InternalError = "InternalError"
)

var (
	commonErrorI18n = map[string]map[string]BaseError{
		InternalError: {
			"zh-CN": {
				ErrorCode:   InternalError,
				Description: "内部错误",
				Solution:    "暂无",
				ErrorLink:   "暂无",
			},
			"en-US": {
				ErrorCode:   InternalError,
				Description: "Internal Server Error",
				Solution:    "None",
				ErrorLink:   "None",
			},
		},
	}
)
