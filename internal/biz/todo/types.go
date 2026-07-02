package todo

import (
	"time"

	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/ordering"
)

// Todo 表示待办事项业务对象。
type Todo struct {
	ID         string    // ID 是待办事项 ULID。
	Title      string    // Title 是待办事项标题。
	Content    string    // Content 是待办事项内容。
	Completed  bool      // Completed 表示待办事项是否完成。
	CreateTime time.Time // CreateTime 是创建时间。
	UpdateTime time.Time // UpdateTime 是更新时间。
}

// ListOption 配置待办事项列表查询。
type ListOption func(*ListOptions)

// ListOptions 表示待办事项列表查询选项。
type ListOptions struct {
	Filter  filtering.Filter // Filter 是 AIP 标准过滤条件。
	OrderBy ordering.OrderBy // OrderBy 是 AIP 标准排序条件。
	Offset  int              // Offset 是分页偏移量。
	Limit   int              // Limit 是分页数量上限。
}
