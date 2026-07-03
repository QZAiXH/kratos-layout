package errors

import (
	"net/http"
	"strings"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
)

const baseContentType = "application"

// ErrorEncoder 编码 HTTP 错误响应并补齐未知错误兜底。
func ErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := kratoserrors.FromError(err)
	if strings.Contains(se.Message, "context deadline exceeded") {
		se = kratoserrors.FromError(ErrTimeout)
	}
	if se.Reason == "" {
		se = kratoserrors.FromError(Unknown(se.Message))
	}

	codec, _ := kratoshttp.CodecForRequest(r, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType(codec.Name()))
	w.WriteHeader(int(se.Code))
	_, _ = w.Write(body)
}

// contentType 返回指定编码的 HTTP Content-Type。
func contentType(subtype string) string {
	return strings.Join([]string{baseContentType, subtype}, "/")
}
