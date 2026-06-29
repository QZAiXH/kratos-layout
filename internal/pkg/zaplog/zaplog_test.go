package zaplog

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMaskPassword 验证日志文本中的密码片段会被脱敏。
func TestMaskPassword(t *testing.T) {
	got := MaskPassword(`login password:"secret" ok`)
	if strings.Contains(got, "secret") || !strings.Contains(got, "password:\"******\"") {
		t.Fatalf("MaskPassword() = %q, want masked password", got)
	}
}

// TestHandlerWritesJSONAndMasksPasswordField 验证 zap 处理器写入 JSON 并脱敏密码字段。
func TestHandlerWritesJSONAndMasksPasswordField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	handler, cleanup, err := NewHandler(WithFilePath(path), WithLevel("debug"))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	logger := slog.New(handler)
	logger.Info("login", slog.String("password", "secret"))
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	output := string(data)
	if !strings.Contains(output, `"msg":"login"`) {
		t.Fatalf("log output = %q, want msg", output)
	}
	if strings.Contains(output, "secret") || !strings.Contains(output, `"password":"******"`) {
		t.Fatalf("log output = %q, want masked password", output)
	}
}
