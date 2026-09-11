package api

import (
	"net/http"

	"github.com/xslasd/goxf/ecode"
)

// ResData 接口统一返回格式
type ResData struct {
	Code int    `json:"code" form:"code"`           // 状态码
	Msg  string `json:"msg" form:"msg"`             // 提示信息
	Data any    `json:"data,omitempty" form:"data"` // 数据
}

type I18nECodeHandler func(err ecode.ECodes) string

type ResOptions struct {
	HTTPStatus  map[int]int
	I18nHandler I18nECodeHandler
}

type ResOption func(*ResOptions)

// WithHTTPStatus 自定义覆盖返回的 HTTP 状态码
func WithHTTPStatus(status map[int]int) ResOption {
	return func(o *ResOptions) {
		o.HTTPStatus = status
	}
}

// WithAddHTTPStatus 增加自定义状态码的 HTTP 映射
func WithAddHTTPStatus(status map[int]int) ResOption {
	return func(o *ResOptions) {
		for k, v := range status {
			o.HTTPStatus[k] = v
		}
	}
}

// WithI18nHandler 注入多语言错误翻译
func WithI18nHandler(handler I18nECodeHandler) ResOption {
	return func(o *ResOptions) {
		o.I18nHandler = handler
	}
}

// CodeToHTTPStatus 全局扩展字典：可供调用方一次性定义好或在初始化时增加自定义状态码的 HTTP 映射
var CodeToHTTPStatus = map[int]int{
	// 成功状态
	ecode.OK.Code(): http.StatusOK,

	// 认证与授权 (401 / 403)
	ecode.AppKeyInvalid.Code():      http.StatusUnauthorized,
	ecode.AccessKeyErr.Code():       http.StatusUnauthorized,
	ecode.SignCheckErr.Code():       http.StatusUnauthorized,
	ecode.Unauthorized.Code():       http.StatusUnauthorized,
	ecode.AccessTokenExpires.Code(): http.StatusUnauthorized,
	ecode.AccessTokenInvalid.Code(): http.StatusUnauthorized,
	ecode.AccessDenied.Code():       http.StatusForbidden,

	// 客户端请求错误 (4xx)
	ecode.RequestErr.Code():       http.StatusBadRequest,
	ecode.NothingFound.Code():     http.StatusNotFound,
	ecode.MethodNotAllowed.Code(): http.StatusMethodNotAllowed,
	ecode.LimitExceed.Code():      http.StatusTooManyRequests,

	// 服务端异常 (5xx)
	ecode.ServerErr.Code():          http.StatusInternalServerError,
	ecode.ServiceUnavailable.Code(): http.StatusServiceUnavailable,
	ecode.Deadline.Code():           http.StatusGatewayTimeout,

	// 通信协议与密钥错误 (WS / JWT / SSE)
	ecode.WSBufferFull.Code():          http.StatusServiceUnavailable,
	ecode.WSClientNotExist.Code():      http.StatusBadRequest,
	ecode.WSClientIdIsNull.Code():      http.StatusBadRequest,
	ecode.WSMessageHandlerIsNil.Code(): http.StatusInternalServerError,
	ecode.SigningKeyIsNull.Code():      http.StatusInternalServerError,
	ecode.SigningKeyLimit.Code():       http.StatusInternalServerError,
	ecode.SSEBufferFull.Code():         http.StatusServiceUnavailable,
	ecode.SSEClientNotExist.Code():     http.StatusBadRequest,
	ecode.SSEClientIdIsNull.Code():     http.StatusBadRequest,
}

// BaseRes 返回结果包装方法
func BaseRes(data any, err error, opts ...ResOption) (int, any) {
	// 浅拷贝全局状态映射字典，避免并发修改或自定义覆盖污染全局配置
	statusMap := make(map[int]int, len(CodeToHTTPStatus))
	for k, v := range CodeToHTTPStatus {
		statusMap[k] = v
	}

	options := &ResOptions{
		HTTPStatus: statusMap,
	}
	for _, opt := range opts {
		opt(options)
	}

	status := http.StatusOK
	var code ecode.ECodes
	if err == nil {
		code = ecode.OK
	} else {
		code = ecode.Cause(err)
	}
	// 查表法动态支持扩充 HTTP 状态码
	if mappedStatus, ok := options.HTTPStatus[code.Code()]; ok {
		status = mappedStatus
	}

	var message string
	if options.I18nHandler != nil {
		message = options.I18nHandler(code)
	}
	if message == "" {
		message = code.Message()
	}
	res := ResData{
		Code: code.Code(),
		Msg:  message,
		Data: data,
	}
	return status, res
}
