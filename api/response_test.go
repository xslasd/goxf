package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xslasd/goxf/ecode"
)

func TestBaseResStatusMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
	}{
		{
			name:       "OK success",
			err:        nil,
			wantStatus: http.StatusOK,
			wantCode:   0,
		},
		{
			name:       "Unauthorized",
			err:        ecode.Unauthorized,
			wantStatus: http.StatusUnauthorized,
			wantCode:   -401,
		},
		{
			name:       "AccessDenied",
			err:        ecode.AccessDenied,
			wantStatus: http.StatusForbidden,
			wantCode:   -403,
		},
		{
			name:       "RequestErr",
			err:        ecode.RequestErr,
			wantStatus: http.StatusBadRequest,
			wantCode:   -400,
		},
		{
			name:       "NothingFound",
			err:        ecode.NothingFound,
			wantStatus: http.StatusNotFound,
			wantCode:   -404,
		},
		{
			name:       "MethodNotAllowed",
			err:        ecode.MethodNotAllowed,
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   -405,
		},
		{
			name:       "LimitExceed",
			err:        ecode.LimitExceed,
			wantStatus: http.StatusTooManyRequests,
			wantCode:   -509,
		},
		{
			name:       "ServerErr",
			err:        ecode.ServerErr,
			wantStatus: http.StatusInternalServerError,
			wantCode:   -500,
		},
		{
			name:       "ServiceUnavailable",
			err:        ecode.ServiceUnavailable,
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   -503,
		},
		{
			name:       "Deadline",
			err:        ecode.Deadline,
			wantStatus: http.StatusGatewayTimeout,
			wantCode:   -504,
		},
		{
			name:       "Raw Error fallback ServerErr",
			err:        errors.New("unexpected database error"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   -500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resData := BaseRes(nil, tt.err)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantCode, resData.Code)
		})
	}
}

func TestBaseResCustomOptions(t *testing.T) {
	// 自定义覆盖状态码
	customStatus := map[int]int{
		-400: http.StatusUnprocessableEntity,
	}
	status, resData := BaseRes("test", ecode.RequestErr, WithAddHTTPStatus(customStatus))
	assert.Equal(t, http.StatusUnprocessableEntity, status)
	assert.Equal(t, -400, resData.Code)
	assert.Equal(t, "test", resData.Data)

	// 全局 CodeToHTTPStatus 未被污染
	assert.Equal(t, http.StatusBadRequest, CodeToHTTPStatus[ecode.RequestErr.Code()])
}
