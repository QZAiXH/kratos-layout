package errors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// TestErrorEncoderFallback 验证普通错误会被统一编码为 UNKNOWN。
func TestErrorEncoderFallback(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)

	ErrorEncoder(recorder, request, errors.New("plain failure"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("状态码 = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(recorder.Body.String(), `"reason":"UNKNOWN"`) {
		t.Fatalf("响应体 = %s, want UNKNOWN reason", recorder.Body.String())
	}
}

// TestUnknownKeepsMessage 验证未知错误保留原始错误消息。
func TestUnknownKeepsMessage(t *testing.T) {
	detail := kratoserrors.FromError(Unknown("plain failure"))
	if detail.Reason != "UNKNOWN" || detail.Message != "plain failure" {
		t.Fatalf("未知错误 = %+v, want UNKNOWN plain failure", detail)
	}
}
