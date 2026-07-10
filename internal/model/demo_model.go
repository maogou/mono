package model

import (
	"gorm.io/gorm"
	"time"
)

type Demo struct {
	ID              uint64     `gorm:"column:id;primaryKey;autoIncrement;unsigned" json:"id"`
	Name            string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Email           string     `gorm:"column:email;type:varchar(255);not null;uniqueIndex:users_email_unique" json:"email"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at;default:null" json:"email_verified_at"`
	Password        string     `gorm:"column:password;type:varchar(255);not null" json:"password"`
	RememberToken   *string    `gorm:"column:remember_token;type:varchar(100);default:null" json:"remember_token"`
	CreatedAt       *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (d *Demo) TableName() string {
	return "demos"
}

// ScopeName 经常使用的查询条件,或者有逻辑判断的查询条件, 麻烦写成scope
func ScopeName(name string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	}
}

// 如果有人要求你不要用id 自增作为主键, 请你取消注释
//func (u *User) BeforeCreate(tx *gorm.DB) error {
//	u.ID = xid.New().String()
//	return nil
//}
