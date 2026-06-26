package zaplog

type Options struct {
	filePath string
	level    string
	rotate   *RotateOptions
}

type RotateOptions struct {
	maxSize    int
	maxBackups int
	maxAge     int
	compress   bool
}

type Option func(*Options)

func WithFilePath(path string) Option {
	return func(o *Options) {
		o.filePath = path
	}
}

func WithLevel(level string) Option {
	return func(o *Options) {
		o.level = level
	}
}

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

func defaultOptions() *Options {
	return &Options{level: "info"}
}
