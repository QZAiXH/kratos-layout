package typecatch

import (
	"fmt"
	"time"

	"github.com/jinzhu/copier"
)

// Option 调整结构体复制选项。
type Option func(*copier.Option)

// WithConverters 追加自定义类型转换器。
func WithConverters(converters ...copier.TypeConverter) Option {
	return func(option *copier.Option) {
		option.Converters = append(option.Converters, converters...)
	}
}

// CopyTo 将同名字段复制到目标结构体。
func CopyTo[SRC, DST any](src *SRC, opts ...Option) (*DST, error) {
	if src == nil {
		return nil, nil
	}
	dst := new(DST)
	option := defaultCopierOption()
	for _, opt := range opts {
		if opt != nil {
			opt(&option)
		}
	}
	if err := copier.CopyWithOption(dst, src, option); err != nil {
		return nil, fmt.Errorf("copy struct: %w", err)
	}
	return dst, nil
}

// defaultCopierOption 返回模板默认复制选项。
func defaultCopierOption() copier.Option {
	return copier.Option{
		DeepCopy: true,
		Converters: []copier.TypeConverter{
			{
				SrcType: (*string)(nil),
				DstType: "",
				Fn: func(src any) (any, error) {
					value, _ := src.(*string)
					if value == nil {
						return "", nil
					}
					return *value, nil
				},
			},
			{
				SrcType: (*time.Time)(nil),
				DstType: time.Time{},
				Fn: func(src any) (any, error) {
					value, _ := src.(*time.Time)
					if value == nil {
						return time.Time{}, nil
					}
					return *value, nil
				},
			},
		},
	}
}
