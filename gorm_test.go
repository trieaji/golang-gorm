package golanggorm

import (
	"testing"
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func OpenConnection() *gorm.DB {
	dialect := mysql.Open("root:@tcp(localhost:3306)/golang_gorm?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := gorm.Open(dialect, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),//memberikan logger
	})
	if err != nil {
		panic(err)
	}

	//gorm juga bisa menggunakan connection pool
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db
}

var db = OpenConnection()

func TestOpenConnection(t *testing.T) {
	assert.NotNil(t, db)
}

func TestExecuteSQL(t *testing.T) {
	err := db.Exec("insert into sample(name) values (?)", "Giyuu").Error//"db.Exec" -> digunakan untuk melakukan manipulasi data
	assert.Nil(t, err)

	err = db.Exec("insert into sample(name) values (?)", "Gojo").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(name) values (?)", "Nanami").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(name) values (?)", "Toji").Error
	assert.Nil(t, err)
}

//how to RawSQL
type Sample struct {
	Id   string
	Name string
}

func TestRawSQL(t *testing.T) {
	var sample Sample
	err := db.Raw("select id, name from sample where id = ?", "1").Scan(&sample).Error
	assert.Nil(t, err)
	assert.Equal(t, "Eko", sample.Name)

	//untuk pengambilan data lebih dari 1
	var samples []Sample
	err = db.Raw("select id, name from sample").Scan(&samples).Error
	assert.Nil(t, err)
	assert.Equal(t, 5, len(samples))
}

//how to using Row
func TestSqlRow(t *testing.T) {
	rows, err := db.Raw("select id, name from sample").Rows()
	assert.Nil(t, err)
	defer rows.Close()

	var samples []Sample
	for rows.Next() {
		var id string
		var name string

		err := rows.Scan(&id, &name)
		assert.Nil(t, err)

		samples = append(samples, Sample{
			Id:   id,
			Name: name,
		})
	}
	assert.Equal(t, 5, len(samples))
}

func TestScanRow(t *testing.T) {//cara lebih mudah untuk melakukan iterasi dibanding TestSqlRow
	rows, err := db.Raw("select id, name from sample").Rows()
	assert.Nil(t, err)
	defer rows.Close()

	var samples []Sample
	for rows.Next() {
		err := db.ScanRows(rows, &samples)
		assert.Nil(t, err)
	}
	assert.Equal(t, 5, len(samples))
}

//create
func TestCreateUser(t *testing.T) {
	user := User{
		Password: "rahasia",
		Name: Name{
			FirstName: "Gojo",
			MiddleName: "Satoru",
			LastName: "Aji",
		},
		Information: "ini akan di ignore",
	}
	response := db.Create(&user)
	assert.Nil(t, response.Error)
	assert.Equal(t, int64(1), response.RowsAffected)
}

//batch insert -> berguna untuk memasukkan data yg banyak dalam sekali input
func TestBatchInsert(t *testing.T) {
	var users []User
	for i := 0; i < 2; i++ {
		users = append(users, User{
			Password: "rahasia",
			Name: Name{
				FirstName: "Laksa ",
			},
		})
	}

	result := db.Create(&users)
	assert.Nil(t, result.Error)
}

//Transaction
func TestTransactionSuccess(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&User{Password: "rahasialah", Name: Name{FirstName: "Nanami"}}).Error
		if err != nil {
			return err
		}

		return nil
	})

	assert.Nil(t, err)
}

func TestTransactionError(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&User{Password: "", Name: Name{FirstName: "Suguru"}}).Error
		if err != nil {
			return err
		}

		return nil
	})

	assert.NotNil(t, err)
}

func TestManualTransactionSuccess(t *testing.T) {
	tx := db.Begin()
	defer tx.Rollback()

	err := tx.Create(&User{Password: "rahasia", Name: Name{FirstName: "Nanami"}}).Error
	assert.Nil(t, err)

	if err == nil {
		tx.Commit()
	}
}

func TestManualTransactionError(t *testing.T) {
	tx := db.Begin()
	defer tx.Rollback()

	err := tx.Create(&User{Password: "", Name: Name{FirstName: "Nanami"}}).Error
	assert.Nil(t, err)

	if err == nil {
		tx.Commit()
	}
}

//query
func TestQuerySingleObject(t *testing.T) {
	user := User{}
	err := db.First(&user).Error //untuk mengambil data yang pertama
	assert.Nil(t, err)
	assert.Equal(t, "1", user.ID)

	// user = User{}
	// err = db.Last(&user).Error //untuk mengambil data yang terakhir
	// assert.Nil(t, err)
	// assert.Equal(t, "9", user.ID)
}

func TestQuerySingleObjectInlineCondition(t *testing.T) {
	user := User{}
	// err := db.First(&user, "id = ?", "5").Error
	err := db.Take(&user, "id = ?", "5").Error
	assert.Nil(t, err)
	assert.Equal(t, "5", user.ID)
	assert.Equal(t, "Nanami", user.Name.FirstName)
}

func TestQueryAllObjects(t *testing.T) {
	var users []User
	err := db.Find(&users, "id in ?", []string{"1", "2", "3", "4"}).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(users))
}

//operator logika untuk query
func TestQueryCondition(t *testing.T) {
	var users []User
	err := db.Where("first_name like ?", "%Gojo%").Where("password = ?", "rahasia").Find(&users).Error//&&
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func TestOrOperator(t *testing.T) {
	var users []User
	err := db.Where("first_name like ?", "%User%").Or("password = ?", "rahasia").Find(&users).Error// |
	assert.Nil(t, err)
	assert.Equal(t, 6, len(users))
}

func TestNotOperator(t *testing.T) {
	var users []User
	err := db.Not("first_name like ?", "%User%").Where("password = ?", "rahasia").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))
}

//select fields -> untuk menentukan kolom mana yg mau diambil
func TestSelectFields(t *testing.T) {
	var users []User
	err := db.Select("id", "first_name").Find(&users).Error
	assert.Nil(t, err)

	for _, user := range users {
		assert.NotNil(t, user.ID)
		assert.NotEqual(t, "", user.Name.FirstName)
	}

	assert.Equal(t, 9, len(users))
}

func TestStructCondition(t *testing.T) {
	userCondition := User{
		Name: Name{
			FirstName: "Gojo",
			// LastName:  "", // tidak bisa, karena dianggap default value(string kosong adalah default value)
		},
		Password: "rahasia",
	}

	var users []User
	err := db.Where(userCondition).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func TestMapCondition(t *testing.T) {
	mapCondition := map[string]interface{}{
		"middle_name": "", //kalau menggunakan map bisa melakukan pencarian string kosong(default value)
		"last_name":   "", //kalau menggunakan map bisa melakukan pencarian string kosong(default value)
	}

	var users []User
	err := db.Where(mapCondition).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 7, len(users))
}

//Order, Limit, Offset -> biasanya digunakan untuk melakukan pagination
func TestOrderLimitOffset(t *testing.T) {
	var users []User
	//Untuk melakukan sorting, kita juga bisa menggunakan method Order()
	//Dan untuk melakukan paging, kita bisa menggunakan method Limit() dan Offset()
	//Offset -> berguna untuk skip misal, ingin skip berapa data
	err := db.Order("id asc, first_name desc").Limit(5).Offset(5).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(users))
}

//query non model
type UserResponse struct {
	ID        int
	FirstName string
	LastName  string
}

func TestQueryNonModel(t *testing.T) {
	var users []UserResponse
	err := db.Model(&User{}).Select("id", "first_name", "last_name").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 9, len(users))
	fmt.Println(users)
}

//update
func TestUpdate(t *testing.T) {//update tanpa memilih kolom untuk di update
	user := User{}
	err := db.Take(&user, "id = ?", "3").Error
	assert.Nil(t, err)

	user.Name.FirstName = "Laksa"
	user.Name.MiddleName = ""
	user.Name.LastName = "Aji"
	user.Password = "rahasia123"

	err = db.Save(&user).Error
	assert.Nil(t, err)
}

func TestUpdateSelectedColumns(t *testing.T) {// update memilih kolom untuk di update
	//menggunakan "Updates map"
	// err := db.Model(&User{}).Where("id = ?", "9").Updates(map[string]interface{}{
	// 	"middle_name": "",
	// 	"last_name":   "Kento",
	// }).Error
	// assert.Nil(t, err)

	//menggunakan "Update"
	// err := db.Model(&User{}).Where("id = ?", "9").Update("last_name", "momonosuke").Error
	// assert.Nil(t, err)

	//menggunakan "Updates struct"
	err := db.Where("id = ?", "9").Updates(User{
		Name: Name{
			FirstName: "Kento",
			LastName:  "Nanami",
		},
	}).Error
	assert.Nil(t, err)
}

//auto increment
func TestAutoIncrement(t *testing.T) {
	for i := 0; i < 10; i++ {
		userLog := UserLog{
			UserId: "3",
			Action: "Test Action3",
		}

		err := db.Create(&userLog).Error
		assert.Nil(t, err)

		assert.NotEqual(t, 0, userLog.ID)
		fmt.Println(userLog.ID)
	}
}

//upsert (update atau insert)
func TestSaveOrUpdate(t *testing.T) {
	userLog := UserLog{
		//ID: , //Tidak set id nya
		UserId: "8",
		Action: "Test Action woy",
	}

	err := db.Save(&userLog).Error // insert //ceritanya tanpa memasukkan id sehingga terjadi create
	assert.Nil(t, err)

	userLog.UserId = "9" //ceritanya memasukkan id sehingga terjadi update
	err = db.Save(&userLog).Error // update
	assert.Nil(t, err)
}

func TestSaveOrUpdateNonAutoIncrement(t *testing.T) { //berguna untuk data yg non auto increment
//ket -> jika ingin melakukan update tapi id nya blum ada di db maka, golang akan melakukan create tapi jika id sudah ada maka, golang akan melakukan update
	user := User{
		ID: 10, //anggaplah id 10 ini belum ada di db (penulisannya harus string -> "10")
		Name: Name{
			FirstName: "Higuruma",
		},
	}

	err := db.Save(&user).Error // insert
	assert.Nil(t, err)

	user.Name.FirstName = "Meliodas"
	err = db.Save(&user).Error // update
	assert.Nil(t, err)
}

func TestConflict(t *testing.T) { //untuk primary key yg duplikat
//ket : jika id belum dibuat maka, golang akan melakukan insert(create) tapi jika ingin create id yg sudah ada maka, golang akan melakukan update
	user := User{
		ID: 10,
		Name: Name{
			FirstName: "Meliodas",
		},
	}

	err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&user).Error // insert
	assert.Nil(t, err)
}

//Delete
func TestDelete(t *testing.T) {
	//cara 1
	var user User //find dulu ke db
	err := db.Take(&user, "id = ?", "4").Error //find dulu ke db
	assert.Nil(t, err)

	err = db.Delete(&user).Error //setelah find db nya maka, lakukan delete
	assert.Nil(t, err)

	//delete secara langsung tanpa melakukan find terlebih dahulu (cara 2)
	err = db.Delete(&User{}, "id = ?", "5").Error
	assert.Nil(t, err)

	//delete secara langsung tanpa melakukan find terlebih dahulu tapi menggunakan "where" (cara 3) 4,5,7 yg kehapus
	err = db.Where("id = ?", "7").Delete(&User{}).Error
	assert.Nil(t, err)
}

//soft delete
func TestSoftDelete(t *testing.T) {
	todo := Todo{
		UserId:      "5",
		Title:       "Todo 5",
		Description: "Description 5",
	}
	err := db.Create(&todo).Error
	assert.Nil(t, err)

	err = db.Delete(&todo).Error
	assert.Nil(t, err)
	assert.NotNil(t, todo.DeletedAt)

	var todos []Todo
	err = db.Find(&todos).Error //datanya tidak bisa diambil karena "deletd_at" nya is NULL kalau ingin memaksa untuk di ambil maka, menggunakan unscoped
	assert.Nil(t, err)
	assert.Equal(t, 0, len(todos))
}

func TestUnscoped(t *testing.T) {
//ket -> Kita kita ingin mengambil data termasuk yang sudah di soft delete, kita bisa gunakan method Unscoped()
//ket -> Method Unscoped() juga bisa digunakan jika kita benar-benar mau melakukan hard delete permanen di database

	// var todo Todo
	// err := db.Unscoped().First(&todo, "id = ?", 1).Error
	// assert.Nil(t, err)
	// fmt.Println(todo)

	// err = db.Unscoped().Delete(&todo).Error
	// assert.Nil(t, err)

	var todos []Todo
	err := db.Unscoped().Find(&todos).Error
	assert.Nil(t, err)
}

//Lock
func TestLock(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		var user User
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&user, "id = ?", "3").Error
		if err != nil {
			return err
		}

		user.Name.FirstName = "Urokodaki"
		user.Name.LastName = "Sakonji"
		err = tx.Save(&user).Error
		return err
	})
	assert.Nil(t, err)
}

//one to one
func TestCreateWallet(t *testing.T) {
	wallet := Wallet{
		UserId:  1,
		Balance: 100,
	}

	err := db.Create(&wallet).Error
	assert.Nil(t, err)
}

func TestRetrieveRelation(t *testing.T) {//untuk mengambil data sekaligus mengambil data yg berelasi. Mirip seperti "join"
	var user User
	err := db.Model(&User{}).Preload("Wallet").Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)

	assert.Equal(t, 1, user.ID)
	assert.Equal(t, 1, user.Wallet.ID)
}

func TestRetrieveRelationJoin(t *testing.T) {
	var user User
	err := db.Model(&User{}).Joins("Wallet").Take(&user, "users.id = ?", "1").Error
	//joins itu lebih cocok untuk relasi one to one
	assert.Nil(t, err)

	assert.Equal(t, 1, user.ID)
	assert.Equal(t, 1, user.Wallet.ID)
	fmt.Println(user)
}

//auto upsert relation -> untuk update atau insert data relasi
func TestAutoCreateUpdate(t *testing.T) {
	user := User{
		// ID:       16,
		Password: "rahasia broo",
		Name: Name{
			FirstName: "User rahasia",
		},
		Wallet: Wallet{ //kalau ada datanya akan melakukan update kalau tidak ada datanya akan melakukan insert
			ID:      5,
			// UserId:  14,
			Balance: 7000,
		},
	}

	err := db.Create(&user).Error
	assert.Nil(t, err)

	//bisa ditambah untuk update wallet yg dituju
}

//skip auto upsert -> untuk skip data relasi yg tidak ingin di update atau insert
func TestSkipAutoCreateUpdate(t *testing.T) {
	user := User{
		// ID:       "21",
		Password: "rahasia",
		Name: Name{
			FirstName: "User 22",
		},
		Wallet: Wallet{
			// ID:      "21",
			// UserId:  "21",
			Balance: 5000000,
		},
	}

	err := db.Omit(clause.Associations).Create(&user).Error
	assert.Nil(t, err)
}

//one to many
func TestUserAndAddresses(t *testing.T) {
	user := User{
		// ID:       "2",
		Password: "rahasia 23",
		Name: Name{
			FirstName: "User 23",
		},
		Wallet: Wallet{
			// ID:      "2",
			// UserId:  "2",
			Balance: 9000,
		},
		Addresses: []Address{
			{
				// UserId:  "23",
				Address: "Jalan A",
			},
			{
				// UserId:  "23",
				Address: "Jalan B",
			},
		},
	}

	err := db.Save(&user).Error
	assert.Nil(t, err)
}

func TestPreloadJoinOneToMany(t *testing.T) {
	var users []User
	err := db.Model(&User{}).Preload("Addresses").Joins("Wallet").Find(&users).Error
	//preload lebih cocok digunakan untuk one to many atau many to many
	assert.Nil(t, err)
}

func TestTakePreloadJoinOneToMany(t *testing.T) {
	var user User
	err := db.Model(&User{}).Preload("Addresses").Joins("Wallet").
		Take(&user, "users.id = ?", "23").Error
	assert.Nil(t, err)
}

//many to one(belongs to)
func TestBelongsTo(t *testing.T) {
	fmt.Println("Preload")
	var addresses []Address
	err := db.Model(&Address{}).Preload("User").Find(&addresses).Error
	assert.Nil(t, err)
	assert.Equal(t, 2, len(addresses))

	fmt.Println("Joins")
	addresses = []Address{}
	err = db.Model(&Address{}).Joins("User").Find(&addresses).Error
	assert.Nil(t, err)
	assert.Equal(t, 2, len(addresses))
}

func TestBelongsToWallet(t *testing.T) {//belongs to (one to one)
	fmt.Println("Preload")
	var wallets []Wallet
	err := db.Model(&Wallet{}).Preload("User").Find(&wallets).Error
	assert.Nil(t, err)

	fmt.Println("Joins")
	wallets = []Wallet{}
	err = db.Model(&Wallet{}).Joins("User").Find(&wallets).Error
	assert.Nil(t, err)
}

//many to many
func TestCreateManyToMany(t *testing.T) {
	// product := Product{ //flexible
	// 	// ID:    "P001",
	// 	Name:  "Contoh Product",
	// 	Price: 1000000,
	// }
	// err := db.Create(&product).Error
	// assert.Nil(t, err)

	err := db.Table("user_like_product").Create(map[string]interface{}{
		"user_id":    9,
		"product_id": 3,
	}).Error
	assert.Nil(t, err)
	
	//kalau misal mau delete/update
	// err = db.Table("user_like_product").Delete(map[string]interface{}{
	// 	"user_id":    "2",
	// 	"product_id": "P001",
	// }).Error
	// assert.Nil(t, err)
}

func TestPreloadManyToManyProduct(t *testing.T) {
	var product Product
	err := db.Preload("LikedByUsers").Take(&product, "id = ?", "1").Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(product.LikedByUsers))
}

func TestPreloadManyToManyUser(t *testing.T) {
	var user User
	err := db.Preload("LikeProducts").Take(&user, "id = ?", "20").Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(user.LikeProducts))
}

//cara mencari relasi(association mode)
func TestAssociationFind(t *testing.T) {
	var product Product
	err := db.Take(&product, "id = ?", "3").Error//coba nanti id nya ganti 3
	assert.Nil(t, err)

	var users []User
	err = db.Model(&product).Where("users.first_name LIKE ?", "Kento%").Association("LikedByUsers").Find(&users)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func TestAssociationAppend(t *testing.T) {
	var user User
	err := db.Take(&user, "id = ?", "2").Error
	assert.Nil(t, err)

	var product Product
	err = db.Take(&product, "id = ?", "3").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Append(&user)//append itu menambahkan
	assert.Nil(t, err)
}

//replace(association mode) -> lebih cocok untuk relasi one to one dalam menggunakan association mode
func TestAssociationReplace(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		var user User
		err := tx.Take(&user, "id = ?", "3").Error
		assert.Nil(t, err)

		wallet := Wallet{
			// ID:      "01",
			UserId:  user.ID,
			Balance: 10000,
		}

		err = tx.Model(&user).Association("Wallet").Replace(&wallet)
		return err
	})
	assert.Nil(t, err)
}

//delete(association mode) -> untuk menghapus relasi
func TestAssociationDelete(t *testing.T) {
	var user User
	err := db.Take(&user, "id = ?", "20").Error
	assert.Nil(t, err)

	var product Product
	err = db.Take(&product, "id = ?", "1").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Delete(&user)
	assert.Nil(t, err)
}

func TestAssociationClear(t *testing.T) {
	var product Product
	err := db.Take(&product, "id = ?", "3").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Clear()
	assert.Nil(t, err)
}

//preloading -> untuk melakukan loading relasi
func TestPreloadingWithCondition(t *testing.T) {
	var user User
	//mau ambil data Wallet nya tapi kalau balance nya lebih dari 100
	//jadi kalau user yang memiliki id "1" itu tidak punya balance sebesar "100" maka, datanya tidak akan diambil
	err := db.Preload("Wallet", "balance > ?", 90).Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)

	fmt.Println(user)
}

//preloading nested
func TestPreloadingNested(t *testing.T) {
	var wallet Wallet
	err := db.Preload("User.Addresses").Take(&wallet, "id = ?", "9").Error
	assert.Nil(t, err)

	fmt.Println(wallet)
	fmt.Println(wallet.User)
	fmt.Println(wallet.User.Addresses)
}

//preload all -> jika kita ingin melakukan preload semua relasi di model
//namun perlu diingat kalau preload all itu tidak melakukan load nested relation
func TestPreloadingAll(t *testing.T) {
	var user User
	err := db.Preload(clause.Associations).Take(&user, "id = ?", "23").Error
	assert.Nil(t, err)
}

//Joins
func TestJoinQuery(t *testing.T) {
	var users []User
	err := db.Joins("join wallets on wallets.user_id = users.id").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 10, len(users))

	//joins manual dan defaultnya adalah left join
	users = []User{}
	err = db.Joins("Wallet").Find(&users).Error // left join
	assert.Nil(t, err)
	assert.Equal(t, 20, len(users))
}

func TestJoinWithCondition(t *testing.T) {
	var users []User
	err := db.Joins("join wallets on wallets.user_id = users.id AND wallets.balance > ?", 500000).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 3, len(users))

	users = []User{}
	err = db.Joins("Wallet").Where("Wallet.balance > ?", 500000).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 3, len(users))
}

//query agregation
func TestCount(t *testing.T) {
	var count int64
	err := db.Model(&User{}).Joins("Wallet").Where("Wallet.balance > ?", 500000).Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, int64(3), count)
}

//other aggregation
type AggregationResult struct {
	TotalBalance int64
	MinBalance   int64
	MaxBalance   int64
	AvgBalance   float64
}

func TestAggregation(t *testing.T) {
	var result AggregationResult
	err := db.Model(&Wallet{}).Select("sum(balance) as total_balance", "min(balance) as min_balance",
		"max(balance) as max_balance", "avg(balance) as avg_balance").Take(&result).Error
	assert.Nil(t, err)

	assert.Equal(t, int64(3043100), result.TotalBalance)
	assert.Equal(t, int64(100), result.MinBalance)
	assert.Equal(t, int64(1000000), result.MaxBalance)
	assert.Equal(t, float64(304310), result.AvgBalance)
}

func TestAggregationGroupByAndHaving(t *testing.T) {
	var results []AggregationResult
	err := db.Model(&Wallet{}).Select("sum(balance) as total_balance", "min(balance) as min_balance",
		"max(balance) as max_balance", "avg(balance) as avg_balance").
		Joins("User").Group("User.id").Having("sum(balance) > ?", 500000).
		Find(&results).Error
	assert.Nil(t, err)
	assert.Equal(t, 3, len(results))
}

//context
func TestContext(t *testing.T) {
	ctx := context.Background()

	var users []User
	err := db.WithContext(ctx).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 20, len(users))
} 

//scopes
func BrokeWalletBalance(db *gorm.DB) *gorm.DB {
	return db.Where("balance = ?", 0)
}

func SultanWalletBalance(db *gorm.DB) *gorm.DB {
	return db.Where("balance > ?", 1000000)
}

func TestScopes(t *testing.T) {
	var wallets []Wallet
	err := db.Scopes(BrokeWalletBalance).Find(&wallets).Error
	assert.Nil(t, err)

	wallets = []Wallet{}
	err = db.Scopes(SultanWalletBalance).Find(&wallets).Error
	assert.Nil(t, err)
}

//migrator -> untuk migrasi(sangat tidak disarankan jika digunakan untuk real case)
func TestMigrator(t *testing.T) {
	err := db.Migrator().AutoMigrate(&GuestBook{})
	assert.Nil(t, err)
}

//hook -> function di dalam Model yang akan dipanggil sebelum melakukan operasi create/query/update/delete
func TestHook(t *testing.T) {
	user := User{
		Password: "rahasia",
		Name: Name{
			FirstName: "User 100",
		},
	}

	err := db.Create(&user).Error
	assert.Nil(t, err)
	assert.NotEqual(t, "", user.ID)

	fmt.Println(user.ID)
}