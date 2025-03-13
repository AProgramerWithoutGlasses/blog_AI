package persistence

import (
	"gorm.io/gorm"

	"grpc-ddd-demo/internal/domain/model/entity"
	"grpc-ddd-demo/internal/domain/repository"
)

type mysqlUserRepository struct {
	db *gorm.DB
}

// NewMySQLUserRepository 返回基于 MySQL 的仓储实现
func NewMySQLUserRepository(db *gorm.DB) repository.UserRepository {
	return &mysqlUserRepository{db: db}
}

func (r *mysqlUserRepository) FindByID(id int64) (*entity.User, error) {
	var user entity.User
	//query := "SELECT id, name, email FROM users WHERE id = ?"
	//row := r.db.QueryRow(query, id)
	//err := row.Scan(&user.ID, &user.Name, &user.Email)
	//if err != nil {
	//	if err == sql.ErrNoRows {
	//		return nil, errors.New("user not found")
	//	}
	//	return nil, err
	//}
	return &user, nil
}

func (r *mysqlUserRepository) Save(user *entity.User) error {
	// 使用 INSERT ... ON DUPLICATE KEY UPDATE 实现更新或插入
	//query := "INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name=?, email=?"
	//_, err := r.db.Exec(query, user.ID, user.Name, user.Email, user.Name, user.Email)
	return nil
}
