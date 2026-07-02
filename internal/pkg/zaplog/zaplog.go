package zaplog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Handler 将 zap 适配为 slog.Handler。
type Handler struct {
	logger *zap.Logger   // logger 是底层 zap 日志器。
	level  zapcore.Level // level 是最小输出级别。
	attrs  []zap.Field   // attrs 是已绑定日志字段。
	groups []string      // groups 是 slog 分组路径。
}

// NewHandler 创建 slog 兼容 zap 处理器。
func NewHandler(opts ...Option) (slog.Handler, func() error, error) {
	options := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	level := zapLevel(options.level)
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	cores := []zapcore.Core{zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)}
	cleanupFile := func() error { return nil }
	if strings.TrimSpace(options.filePath) != "" {
		writer, cleanup, err := fileWriter(options)
		if err != nil {
			return nil, nil, err
		}
		cleanupFile = cleanup
		cores = append(cores, zapcore.NewCore(encoder, writer, level))
	}

	logger := zap.New(zapcore.NewTee(cores...))
	return &Handler{logger: logger, level: level}, func() error {
		if err := logger.Sync(); err != nil {
			text := err.Error()
			if !strings.Contains(text, "invalid argument") && !strings.Contains(text, "bad file descriptor") {
				return err
			}
		}
		return cleanupFile()
	}, nil
}

// Enabled 判断指定级别是否需要输出。
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return h.level.Enabled(zapLevelFromSlog(level))
}

// Handle 将 slog 记录写入 zap。
func (h *Handler) Handle(_ context.Context, record slog.Record) error {
	fields := append([]zap.Field(nil), h.attrs...)
	if src := record.Source(); src != nil {
		fields = append(fields, zap.String("caller", fmt.Sprintf("%s:%d", filepath.Base(src.File), src.Line)))
	}
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, attrFields(h.groups, attr)...)
		return true
	})
	h.logger.Log(zapLevelFromSlog(record.Level), record.Message, fields...)
	return nil
}

// WithAttrs 返回绑定属性后的处理器。
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := h.clone()
	for _, attr := range attrs {
		next.attrs = append(next.attrs, attrFields(next.groups, attr)...)
	}
	return next
}

// WithGroup 返回绑定分组后的处理器。
func (h *Handler) WithGroup(name string) slog.Handler {
	if strings.TrimSpace(name) == "" {
		return h
	}
	next := h.clone()
	next.groups = append(next.groups, name)
	return next
}

// clone 复制处理器状态。
func (h *Handler) clone() *Handler {
	next := *h
	next.attrs = append([]zap.Field(nil), h.attrs...)
	next.groups = append([]string(nil), h.groups...)
	return &next
}

// attrFields 将 slog 属性转换为 zap 字段。
func attrFields(groups []string, attr slog.Attr) []zap.Field {
	attr.Value = attr.Value.Resolve()
	if attr.Key == "" {
		return nil
	}
	key := strings.Join(append(append([]string(nil), groups...), attr.Key), ".")
	if isSecretKey(attr.Key) {
		return []zap.Field{zap.String(key, "******")}
	}
	switch attr.Value.Kind() {
	case slog.KindString:
		value := attr.Value.String()
		if attr.Key == "args" {
			value = MaskPassword(value)
		}
		return []zap.Field{zap.String(key, value)}
	case slog.KindBool:
		return []zap.Field{zap.Bool(key, attr.Value.Bool())}
	case slog.KindDuration:
		return []zap.Field{zap.Duration(key, attr.Value.Duration())}
	case slog.KindFloat64:
		return []zap.Field{zap.Float64(key, attr.Value.Float64())}
	case slog.KindInt64:
		return []zap.Field{zap.Int64(key, attr.Value.Int64())}
	case slog.KindUint64:
		return []zap.Field{zap.Uint64(key, attr.Value.Uint64())}
	case slog.KindTime:
		return []zap.Field{zap.Time(key, attr.Value.Time())}
	case slog.KindGroup:
		var fields []zap.Field
		for _, child := range attr.Value.Group() {
			fields = append(fields, attrFields(append(append([]string(nil), groups...), attr.Key), child)...)
		}
		return fields
	default:
		return []zap.Field{zap.Any(key, attr.Value.Any())}
	}
}

// fileWriter 根据配置创建日志文件 writer。
func fileWriter(options *Options) (zapcore.WriteSyncer, func() error, error) {
	if err := os.MkdirAll(filepath.Dir(options.filePath), os.ModePerm); err != nil {
		return nil, nil, err
	}
	if options.rotate != nil {
		writer := &lumberjack.Logger{
			Filename:   options.filePath,
			MaxSize:    options.rotate.maxSize,
			MaxBackups: options.rotate.maxBackups,
			MaxAge:     options.rotate.maxAge,
			Compress:   options.rotate.compress,
		}
		return zapcore.AddSync(writer), writer.Close, nil
	}
	file, err := os.OpenFile(options.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}
	return zapcore.AddSync(file), file.Close, nil
}

// zapLevel 将配置字符串转换为 zap 级别。
func zapLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zap.DebugLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

// zapLevelFromSlog 将 slog 级别转换为 zap 级别。
func zapLevelFromSlog(level slog.Level) zapcore.Level {
	switch {
	case level < slog.LevelInfo:
		return zap.DebugLevel
	case level < slog.LevelWarn:
		return zap.InfoLevel
	case level < slog.LevelError:
		return zap.WarnLevel
	default:
		return zap.ErrorLevel
	}
}

// MaskPassword 脱敏日志文本中的 password 值。
func MaskPassword(input string) string {
	re := regexp.MustCompile(`(?i)(password["=: ]+)([^",}\s]+)`)
	return re.ReplaceAllString(input, `${1}******`)
}

// isSecretKey 判断字段名是否为敏感字段。
func isSecretKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "password") || strings.Contains(key, "secret") || strings.Contains(key, "token")
}

var _ slog.Handler = (*Handler)(nil)
