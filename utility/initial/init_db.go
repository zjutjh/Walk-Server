package initial

import (
	"fmt"
	"os"
	"walk-server/global"
	"walk-server/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func DBInit() {
	// 从配置文件中读取数据库信息
	dbHost := global.Config.GetString("database.host")
	dbUser := global.Config.GetString("database.user")
	dbPassport := global.Config.GetString("database.passport")
	dbPort := global.Config.GetString("database.port")
	dbName := global.Config.GetString("database.name")

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=true",
		dbUser, dbPassport, dbHost, dbPort, dbName)

	var err error
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true, // 开启预编译
	})
	if err != nil {
		fmt.Println("数据库连接错误")
		fmt.Println(err)
		os.Exit(-1)
	}

	// 这个地方需要填入要迁移的表
	err = global.DB.AutoMigrate(&model.Person{}, &model.Team{}, &model.Message{})
	if err != nil {
		fmt.Println("数据表创建错误")
		os.Exit(-1)
	}
}
