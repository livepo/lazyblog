package invoker

import (
	"fmt"
	"lazyblog/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func Init() {
	// init db
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
		config.Cfg.Mysql.User,
		config.Cfg.Mysql.Password,
		config.Cfg.Mysql.Host,
		config.Cfg.Mysql.Port,
		config.Cfg.Mysql.Database,
	)
	fmt.Println("dsn:", dsn)
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = database
}
