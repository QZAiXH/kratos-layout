package dep

import (
	"testing"
	"time"

	"helloworld/internal/conf"

	"google.golang.org/protobuf/types/known/durationpb"
)

// TestNewDBAllowsEmptyConfig 验证空数据库配置不会创建外部连接。
func TestNewDBAllowsEmptyConfig(t *testing.T) {
	db, cleanup, err := NewDB(&conf.Data{})
	if err != nil {
		t.Fatalf("NewDB(empty) error = %v", err)
	}
	defer cleanup()

	if db != nil {
		t.Fatalf("NewDB(empty) = %v, want nil", db)
	}
}

// TestNewDBOpensConfiguredClient 验证数据库配置会创建 SQL 客户端。
func TestNewDBOpensConfiguredClient(t *testing.T) {
	db, cleanup, err := NewDB(&conf.Data{
		Database: &conf.Data_Database{
			Driver: "mysql",
			Source: "root:root@tcp(127.0.0.1:3306)/test?parseTime=True&loc=Local",
		},
	})
	if err != nil {
		t.Fatalf("NewDB(configured) error = %v", err)
	}
	defer cleanup()

	if db == nil {
		t.Fatal("NewDB(configured) = nil, want client")
	}
}

// TestNewDBRejectsUnknownDriver 验证未知数据库驱动会在启动初始化阶段报错。
func TestNewDBRejectsUnknownDriver(t *testing.T) {
	_, _, err := NewDB(&conf.Data{
		Database: &conf.Data_Database{
			Driver: "unknown",
			Source: "ignored",
		},
	})
	if err == nil {
		t.Fatal("NewDB(unknown driver) error = nil, want error")
	}
}

// TestNewRedisAllowsEmptyConfig 验证空 Redis 配置不会创建客户端。
func TestNewRedisAllowsEmptyConfig(t *testing.T) {
	client, cleanup, err := NewRedis(&conf.Data{})
	if err != nil {
		t.Fatalf("NewRedis(empty) error = %v", err)
	}
	defer cleanup()

	if client != nil {
		t.Fatalf("NewRedis(empty) = %v, want nil", client)
	}
}

// TestNewRedisCreatesConfiguredClient 验证 Redis 配置会创建带对应选项的客户端。
func TestNewRedisCreatesConfiguredClient(t *testing.T) {
	client, cleanup, err := NewRedis(&conf.Data{
		Redis: &conf.Data_Redis{
			Network:      "tcp",
			Addr:         "127.0.0.1:6379",
			Password:     "secret",
			ReadTimeout:  durationpb.New(200 * time.Millisecond),
			WriteTimeout: durationpb.New(300 * time.Millisecond),
		},
	})
	if err != nil {
		t.Fatalf("NewRedis(configured) error = %v", err)
	}
	defer cleanup()

	if client == nil {
		t.Fatal("NewRedis(configured) = nil, want client")
	}

	options := client.Options()
	if options.Network != "tcp" || options.Addr != "127.0.0.1:6379" || options.Password != "secret" {
		t.Fatalf("Redis options = %+v, want configured network/address/password", options)
	}
	if options.ReadTimeout != 200*time.Millisecond || options.WriteTimeout != 300*time.Millisecond {
		t.Fatalf("Redis timeout = %s/%s, want 200ms/300ms", options.ReadTimeout, options.WriteTimeout)
	}
}
