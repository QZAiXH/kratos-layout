package main

import (
	"os"
	"strings"

	"github.com/QZAiXH/kratos-layout/internal/conf"
)

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

func setString(key string, apply func(string)) {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		apply(value)
	}
}

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

func ensureAuth(bc *conf.Bootstrap) *conf.Auth {
	if bc.Auth == nil {
		bc.Auth = &conf.Auth{}
	}
	return bc.Auth
}

func ensureLog(bc *conf.Bootstrap) *conf.Log {
	if bc.Log == nil {
		bc.Log = &conf.Log{}
	}
	return bc.Log
}
