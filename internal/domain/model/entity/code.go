package entity

import (
	"gorm.io/gorm"
	"siwuai/internal/domain/model/dto"
)

// Code 代表代码存储表
type Code struct {
	gorm.Model
	Key         string
	Question    string
	Explanation string
	// 一对多关联，一个 Code 可以有多个 History 记录
	Histories []History `gorm:"foreignKey:CodeID"`
}

// History 代表历史记录表
type History struct {
	gorm.Model
	UserID uint
	CodeID uint // 外键，关联 Code 表的 ID
	// 可选：如果需要建立反向关联，可加上 Code 字段
	Code Code `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (c Code) CodeToDto() *dto.Code {
	return &dto.Code{
		ID:          c.ID,
		Question:    c.Question,
		Explanation: c.Explanation,
		Key:         c.Key,
	}
}

func (c Code) DtoToCode(dto *dto.Code) *Code {
	return &Code{
		Model: gorm.Model{
			ID: c.ID,
		},
		Key:         dto.Key,
		Question:    dto.Question,
		Explanation: dto.Explanation,
	}
}
