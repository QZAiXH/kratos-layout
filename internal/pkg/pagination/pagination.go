package pagination

const (
	DefaultPageSize = 20
	MaxPageSize     = 200
)

// Normalize 规范化页码参数并返回 limit 与 offset。
func Normalize(pageSize, pageNumber int32) (limit, offset int) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	if pageNumber <= 0 {
		pageNumber = 1
	}
	return int(pageSize), int((pageNumber - 1) * pageSize)
}
