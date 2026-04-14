package api

import (
	"net/http"

	"github.com/xslasd/goxf/ecode"
)

// ListRes 分页接口统一返回参数
type ListRes struct {
	Rows  any   `json:"rows"`
	Total int64 `json:"total"`
}

// TreeRes 树形接口统一返回参数
type TreeRes struct {
	Tree any `json:"tree"`
}

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
	-400: http.StatusBadRequest,
	-500: http.StatusInternalServerError,
	-2:   http.StatusUnauthorized,
	-401: http.StatusUnauthorized,
	-658: http.StatusUnauthorized,
	-403: http.StatusForbidden,
	-406: http.StatusForbidden,
	-407: http.StatusForbidden,
	-404: http.StatusNotFound,
}

// BaseRes 返回结果包装方法
func BaseRes(data any, err error, opts ...ResOption) (int, any) {
	// 解析自定义参数
	options := &ResOptions{
		HTTPStatus: CodeToHTTPStatus,
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
