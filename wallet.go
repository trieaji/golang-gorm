package golanggorm

import "time"

type Wallet struct {
	ID        int    	`gorm:"primary_key;column:id"`
	UserId    int    	`gorm:"column:user_id"`
	Balance   int64     `gorm:"column:balance"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	User      *User     `gorm:"foreignKey:user_id;references:id"`//relasi belongs to (one to one)
}

func (w *Wallet) TableName() string {
	return "wallets"
}