package ent

import (
	stdsql "database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

// SQLDB 返回当前 Database 绑定的底层 SQL 连接池。
func (db *Database) SQLDB() (*stdsql.DB, bool) {
	if db == nil || db.client == nil {
		return nil, false
	}
	return sqlDBFromDriver(db.client.driver)
}

// Dialect 返回当前 Database 使用的 SQL 方言。
func (db *Database) Dialect() string {
	if db == nil || db.client == nil || db.client.driver == nil {
		return ""
	}
	return dialectFromDriver(db.client.driver)
}

// sqlDBFromDriver 从 Ent driver 中提取底层 SQL 连接池。
func sqlDBFromDriver(driver dialect.Driver) (*stdsql.DB, bool) {
	switch d := driver.(type) {
	case *entsql.Driver:
		return d.DB(), true
	case *dialect.DebugDriver:
		return sqlDBFromDriver(d.Driver)
	default:
		return nil, false
	}
}

// dialectFromDriver 从 Ent driver 中提取 SQL 方言。
func dialectFromDriver(driver dialect.Driver) string {
	switch d := driver.(type) {
	case *dialect.DebugDriver:
		return dialectFromDriver(d.Driver)
	default:
		return driver.Dialect()
	}
}
