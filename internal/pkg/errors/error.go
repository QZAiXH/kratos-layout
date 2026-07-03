package errors

import kratoserrors "github.com/go-kratos/kratos/v3/errors"

var (
	// ErrTimeout 表示请求处理超时。
	ErrTimeout = kratoserrors.New(504, "TIMEOUT", "")
)

// Unknown 创建未知错误。
func Unknown(msg string) error {
	return kratoserrors.New(500, "UNKNOWN", msg)
}
