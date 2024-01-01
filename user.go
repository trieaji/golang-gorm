package golanggorm

import "time"

// User => users (contoh penamaan tabel akan dimapping secara otomatis oleh gorm)
// OrderDetail => order_details (contoh penamaan tabel akan dimapping secara otomatis oleh gorm)
// Order => orders (contoh penamaan tabel akan dimapping secara otomatis oleh gorm)

//ini adalah cara pembuatan model atau entity
type User struct {
	ID           int		`gorm:"primary_key;column:id;autoIncrement"`
	Password     string		`gorm:"column:password"`
	Name         Name		`gorm:"embedded"` //embedded strcut
	CreatedAt    time.Time	`gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time	`gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	Information  string    	`gorm:"-"`//artinya tidak ada di db 
	Wallet       Wallet    `gorm:"foreignKey:user_id;references:id"` //one to one (jngan lupa datanya dikasih unique)
	Addresses    []Address `gorm:"foreignKey:user_id;references:id"` //one to many
	LikeProducts []Product `gorm:"many2many:user_like_product;foreignKey:id;joinForeignKey:user_id;references:id;joinReferences:product_id"`
}

//cara mengubah nama table mapping
func(u *User) TableName() string {
	return "users"
}

//hook
// func (u *User) BeforeCreate(db *gorm.DB) error {
// 	if u.ID == 0 {
// 		u.ID = "user-" + time.Now().Format("20060102150405")
// 	}
// 	return nil
// }

//embedded struct -> berguna untuk meng grouping filed-field yang terlalu banyak di struct
type Name struct {
	FirstName  string `gorm:"column:first_name"`
	MiddleName string `gorm:"column:middle_name"`
	LastName   string `gorm:"column:last_name"`
}

type UserLog struct {
	ID        int    	`gorm:"primary_key;column:id;autoIncrement"`
	UserId    string 	`gorm:"column:user_id"`
	Action    string 	`gorm:"column:action"`
	CreatedAt int64 	`gorm:"column:created_at;autoCreateTime:milli"` //milli adalah timestamp tracking (mengubah waktunya menjadi millisecond)
	UpdatedAt int64 	`gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"` //milli adalah timestamp tracking (mengubah waktunya menjadi millisecond)
}

func (l *UserLog) TableName() string {
	return "user_logs"
}