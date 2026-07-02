package todo

import (
	"sync"

	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"
	"github.com/QZAiXH/kratos-layout/internal/data/base"
)

// todoRepo 用内存存储实现待办事项仓储。
type todoRepo struct {
	data *base.Data // data 是数据层共享依赖。

	mu    sync.RWMutex             // mu 保护内存待办事项集合。
	todos map[string]*todobiz.Todo // todos 保存待办事项快照。
}

// NewRepo 创建待办事项仓储实例。
func NewRepo(data *base.Data) todobiz.Repo {
	return &todoRepo{
		data:  data,
		todos: make(map[string]*todobiz.Todo),
	}
}
