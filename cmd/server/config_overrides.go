package main

import (
	"os"
	"strings"

	"github.com/QZAiXH/kratos-layout/internal/conf"
)

// applyEnvOverrides 将环境变量覆盖到启动配置。
func applyEnvOverrides(bc *conf.Bootstrap) {
	if bc == nil {
		return
	}
	setString("APP_DATA_DATABASE_SOURCE", func(v string) {
		ensureData(bc).Database.Source = v
	})
	setString("APP_DATA_REDIS_ADDR", func(v string) {
		ensureData(bc).Redis.Addr = v
	})
	setString("APP_DATA_REDIS_PASSWORD", func(v string) {
		ensureData(bc).Redis.Password = v
	})
	setString("APP_AUTH_PRIVATE_KEY_PATH", func(v string) {
		ensureAuth(bc).PrivateKeyPath = v
	})
	setString("APP_LOG_LEVEL", func(v string) {
		ensureLog(bc).Level = v
	})
}

// setString 读取非空字符串环境变量并应用到配置。
func setString(key string, apply func(string)) {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		apply(value)
	}
}

// ensureData 确保 data 配置及其子配置已初始化。
func ensureData(bc *conf.Bootstrap) *conf.Data {
	if bc.Data == nil {
		bc.Data = &conf.Data{}
	}
	if bc.Data.Database == nil {
		bc.Data.Database = &conf.Data_Database{}
	}
	if bc.Data.Redis == nil {
		bc.Data.Redis = &conf.Data_Redis{}
	}
	return bc.Data
}

// ensureAuth 确保 auth 配置已初始化。
func ensureAuth(bc *conf.Bootstrap) *conf.Auth {
	if bc.Auth == nil {
		bc.Auth = &conf.Auth{}
	}
	return bc.Auth
}

// ensureLog 确保 log 配置已初始化。
func ensureLog(bc *conf.Bootstrap) *conf.Log {
	if bc.Log == nil {
		bc.Log = &conf.Log{}
	}
	return bc.Log
}
