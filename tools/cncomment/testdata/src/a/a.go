package a

// ExportedFunc 执行业务动作。
func ExportedFunc() {}

func MissingFunc() {} // want ".*"

type receiver struct{}

// OK 执行业务动作。
func (receiver) OK() {}

func (receiver) MissingMethod() { // want ".*"
}

// ExportedStruct 保存业务数据。
type ExportedStruct struct {
	Name   string // want ".*"
	Title  string // 标题。
	hidden string
}

type MissingStruct struct { // want ".*"
	OK string // 字段。
}

// ExportedInterface 表示业务接口。
type ExportedInterface interface {
	Do() // want ".*"
	OK() // 操作。
}

type MissingInterface interface { // want ".*"
	OK() // 操作。
}

func hidden() {}

type hiddenStruct struct {
	Exported string // want ".*"
}
