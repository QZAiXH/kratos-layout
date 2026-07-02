package pagination

const (
	DefaultPageSize = 20
	MaxPageSize     = 200
)

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
