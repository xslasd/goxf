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
		if o.HTTPStatus == nil {
			o.HTTPStatus = make(map[int]int, len(status))
		}
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

// BaseRes 返回结果包装方法 (兼顾零分配极致性能与层叠自定义隔离)
func BaseRes(data any, err error, opts ...ResOption) (int, ResData) {
	code := ecode.OK
	if err != nil {
		code = ecode.Cause(err)
	}
	codeVal := code.Code()

	// 快速路径：99% 的请求无 opts，直接查只读全局字典，零堆内存分配
	if len(opts) == 0 {
		status := http.StatusOK
		if mapped, ok := CodeToHTTPStatus[codeVal]; ok {
			status = mapped
		}
		return status, ResData{
			Code: codeVal,
			Msg:  code.Message(),
			Data: data,
		}
	}

	// 慢路径：存在自定义 options 时按需解析
	options := &ResOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 状态码两级查找：局部自定义字典优先，未命中回退全局只读字典，避免全量拷贝
	status := http.StatusOK
	if options.HTTPStatus != nil {
		if mapped, ok := options.HTTPStatus[codeVal]; ok {
			status = mapped
		} else if globalMapped, ok := CodeToHTTPStatus[codeVal]; ok {
			status = globalMapped
		}
	} else if globalMapped, ok := CodeToHTTPStatus[codeVal]; ok {
		status = globalMapped
	}

	var message string
	if options.I18nHandler != nil {
		message = options.I18nHandler(code)
	}
	if message == "" {
		message = code.Message()
	}

	return status, ResData{
		Code: codeVal,
		Msg:  message,
		Data: data,
	}
}
