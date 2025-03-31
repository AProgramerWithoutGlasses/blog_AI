package persistence

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/infrastructure/config"
)

// NewMySQLDB 根据配置文件创建并初始化 GORM 的 MySQL 数据库连接
func NewMySQLDB(cfg config.Config) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MySQL.Username,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DBName,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 取消外键约束
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "sw_ai_", // 设置表前缀
		},
	})
	if err != nil {
		err = fmt.Errorf("gorm.Open() err: %v", err)
		return
	}

	// 表迁移
	err = db.AutoMigrate(
		&entity.Code{},
		&entity.History{},
		&entity.Article{},
	)
	if err != nil {
		err = fmt.Errorf("db.AutoMigrate() err: %v", err)
		return
	}

	fmt.Println("MySQL 连接成功")

	return
}
