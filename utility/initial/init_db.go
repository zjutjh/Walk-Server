package initial

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"walk-server/model"
)

var DB *gorm.DB

func DBInit() {
	// 从配置文件中读取数据库信息
	dbHost := Config.GetString("database.host")
	dbUser := Config.GetString("database.user")
	dbPassport := Config.GetString("database.passport")
	dbPort := Config.GetString("database.port")
	dbName := Config.GetString("database.name")

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v",
		dbUser, dbPassport, dbHost, dbPort, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("数据库连接错误")
		fmt.Println(err)
		os.Exit(-1)
	}

	err = DB.AutoMigrate(&model.Person{}, &model.Team{}, &model.TeamCount{})
	if err != nil {
		fmt.Println("数据表创建错误")
		os.Exit(-1)
	}
}
