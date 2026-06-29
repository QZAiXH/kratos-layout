package typecatch

import (
	"testing"
	"time"

	"github.com/jinzhu/copier"
)

type copySrc struct {
	Nickname *string
	When     *time.Time
	Status   int8
}

type copyDst struct {
	Nickname string
	When     time.Time
	Status   int32
}

// TestCopyToDTO 验证同名字段复制会处理空指针和基础类型转换。
func TestCopyToDTO(t *testing.T) {
	dst, err := CopyToDTO[copySrc, copyDst](nil)
	if err != nil {
		t.Fatalf("CopyToDTO(nil) error = %v", err)
	}
	if dst != nil {
		t.Fatalf("CopyToDTO(nil) = %+v, want nil", dst)
	}

	dst, err = CopyToDTO[copySrc, copyDst](&copySrc{Status: 7})
	if err != nil {
		t.Fatalf("CopyToDTO() error = %v", err)
	}
	if dst.Nickname != "" || !dst.When.IsZero() || dst.Status != 7 {
		t.Fatalf("CopyToDTO() = %+v, want zero pointer fields and copied status", dst)
	}
}

// TestCopyToDTOWithFieldNameMapping 验证字段名映射能复制语义相同但名称不同的字段。
func TestCopyToDTOWithFieldNameMapping(t *testing.T) {
	type mappingSrc struct {
		UserID string
	}
	type mappingDst struct {
		ID string
	}

	dst, err := CopyToDTO[mappingSrc, mappingDst](&mappingSrc{UserID: "user_1"}, WithFieldNameMapping(copier.FieldNameMapping{
		SrcType: mappingSrc{},
		DstType: mappingDst{},
		Mapping: map[string]string{"UserID": "ID"},
	}))
	if err != nil {
		t.Fatalf("CopyToDTO() error = %v", err)
	}
	if dst.ID != "user_1" {
		t.Fatalf("CopyToDTO() ID = %q, want user_1", dst.ID)
	}
}

// TestCopySliceToDTO 验证切片复制会保留空元素并转换普通元素。
func TestCopySliceToDTO(t *testing.T) {
	nickname := "A"
	dst, err := CopySliceToDTO[copySrc, copyDst]([]*copySrc{
		{Nickname: &nickname, Status: 1},
		nil,
	})
	if err != nil {
		t.Fatalf("CopySliceToDTO() error = %v", err)
	}
	if len(dst) != 2 || dst[0].Nickname != "A" || dst[0].Status != 1 || dst[1] != nil {
		t.Fatalf("CopySliceToDTO() = %+v, want copied item and nil item", dst)
	}
}
