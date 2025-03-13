package persistence

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"grpc-ddd-demo/internal/infrastructure/config"
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
		fmt.Println("persistence.NewMySQLDB() gorm.Open() err: ", err)
		return
	}
	fmt.Println("MySQL 连接成功")

	// 表迁移
	//err = db.AutoMigrate(&entity.Code{}, &entity.History{})
	//if err != nil {
	//	fmt.Println("persistence.NewMySQLDB() db.AutoMigrate() err: ", err)
	//	return
	//}
	return
}
