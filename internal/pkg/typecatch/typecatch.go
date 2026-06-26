package typecatch

import (
	"fmt"
	"time"

	"github.com/jinzhu/copier"
)

type Option func(*copier.Option)

func WithConverters(converters ...copier.TypeConverter) Option {
	return func(option *copier.Option) {
		option.Converters = append(option.Converters, converters...)
	}
}

func WithFieldNameMapping(mappings ...copier.FieldNameMapping) Option {
	return func(option *copier.Option) {
		option.FieldNameMapping = append(option.FieldNameMapping, mappings...)
	}
}

func CopyToDTO[SRC, DST any](src *SRC, opts ...Option) (*DST, error) {
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
		return nil, fmt.Errorf("copy dto: %w", err)
	}
	return dst, nil
}

func CopySliceToDTO[SRC, DST any](src []*SRC, opts ...Option) ([]*DST, error) {
	if src == nil {
		return nil, nil
	}

	option := defaultCopierOption()
	for _, opt := range opts {
		if opt != nil {
			opt(&option)
		}
	}

	dst := make([]*DST, len(src))
	for i, item := range src {
		if item == nil {
			continue
		}
		out := new(DST)
		if err := copier.CopyWithOption(out, item, option); err != nil {
			return nil, fmt.Errorf("copy dto slice: %w", err)
		}
		dst[i] = out
	}
	return dst, nil
}

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
