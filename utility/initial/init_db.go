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

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
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

	// 初始化 teamCount 表的数据
	var teamCount model.TeamCount
	for i := 0; i <= 3; i++ { // 枚举天数
		for j := 1; j <= 5; j++ { // 枚举路线编号
			result := DB.Where("day_campus = ?", i*10+j).Take(&model.TeamCount{})
			if result.RowsAffected == 0 {
				teamCount.DayCampus = uint8(i*10 + j)
				teamCount.Count = 0
				DB.Create(&teamCount) // 创建队伍计数器
			}
		}
	}
}
