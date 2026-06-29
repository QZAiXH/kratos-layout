package data

import (
	"testing"
	"time"

	"helloworld/internal/conf"

	"google.golang.org/protobuf/types/known/durationpb"
)

// TestNewDataAllowsEmptyConfig 验证空配置不会创建外部连接客户端。
func TestNewDataAllowsEmptyConfig(t *testing.T) {
	data, cleanup, err := NewData(&conf.Data{})
	if err != nil {
		t.Fatalf("NewData(empty) error = %v", err)
	}
	defer cleanup()

	if data.DB != nil || data.RDB != nil {
		t.Fatalf("NewData(empty) = %+v, want nil clients", data)
	}
}

// TestNewDataOpensConfiguredClients 验证数据库和 Redis 配置会创建对应客户端。
func TestNewDataOpensConfiguredClients(t *testing.T) {
	data, cleanup, err := NewData(&conf.Data{
		Database: &conf.Data_Database{
			Driver: "mysql",
			Source: "root:root@tcp(127.0.0.1:3306)/test?parseTime=True&loc=Local",
		},
		Redis: &conf.Data_Redis{
			Network:      "tcp",
			Addr:         "127.0.0.1:6379",
			Password:     "secret",
			ReadTimeout:  durationpb.New(200 * time.Millisecond),
			WriteTimeout: durationpb.New(300 * time.Millisecond),
		},
	})
	if err != nil {
		t.Fatalf("NewData(configured) error = %v", err)
	}
	defer cleanup()

	if data.DB == nil {
		t.Fatal("NewData(configured) DB = nil, want client")
	}
	if data.RDB == nil {
		t.Fatal("NewData(configured) RDB = nil, want client")
	}

	options := data.RDB.Options()
	if options.Network != "tcp" || options.Addr != "127.0.0.1:6379" || options.Password != "secret" {
		t.Fatalf("Redis options = %+v, want configured network/address/password", options)
	}
	if options.ReadTimeout != 200*time.Millisecond || options.WriteTimeout != 300*time.Millisecond {
		t.Fatalf("Redis timeout = %s/%s, want 200ms/300ms", options.ReadTimeout, options.WriteTimeout)
	}
}

// TestNewDataRejectsUnknownDatabaseDriver 验证未知数据库驱动会在启动初始化阶段报错。
func TestNewDataRejectsUnknownDatabaseDriver(t *testing.T) {
	_, _, err := NewData(&conf.Data{
		Database: &conf.Data_Database{
			Driver: "unknown",
			Source: "ignored",
		},
	})
	if err == nil {
		t.Fatal("NewData(unknown driver) error = nil, want error")
	}
}
