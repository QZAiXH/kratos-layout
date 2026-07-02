package zaplog

// Options 表示 zap slog 处理器配置。
type Options struct {
	filePath string         // filePath 是日志文件路径。
	level    string         // level 是日志级别。
	rotate   *RotateOptions // rotate 是日志轮转配置。
}

// RotateOptions 表示日志文件轮转配置。
type RotateOptions struct {
	maxSize    int  // maxSize 是单个日志文件最大 MB。
	maxBackups int  // maxBackups 是保留旧文件数量。
	maxAge     int  // maxAge 是保留旧文件天数。
	compress   bool // compress 表示是否压缩旧文件。
}

// Option 修改日志处理器配置。
type Option func(*Options)

// WithFilePath 设置日志文件路径。
func WithFilePath(path string) Option {
	return func(o *Options) {
		o.filePath = path
	}
}

// WithLevel 设置日志级别。
func WithLevel(level string) Option {
	return func(o *Options) {
		o.level = level
	}
}

// WithRotate 设置日志轮转参数。
func WithRotate(maxSize, maxBackups, maxAge int, compress bool) Option {
	return func(o *Options) {
		o.rotate = &RotateOptions{
			maxSize:    maxSize,
			maxBackups: maxBackups,
			maxAge:     maxAge,
			compress:   compress,
		}
	}
}

// defaultOptions 返回默认日志配置。
func defaultOptions() *Options {
	return &Options{level: "info"}
}
