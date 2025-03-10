package persistence

import (
	"database/sql"
	"fmt"
	"log"

	"grpc-ddd-demo/internal/infrastructure/config"

	_ "github.com/go-sql-driver/mysql"
)

// NewMySQLDB 根据配置文件创建并初始化 MySQL 数据库连接
func NewMySQLDB(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=%v",
		cfg.MySQL.Username,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DBName,
		cfg.MySQL.ParseTime,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open error: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping error: %w", err)
	}
	log.Println("MySQL 连接成功")
	return db, nil
}
