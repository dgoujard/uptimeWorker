package gorm

import (
	"fmt"
	"github.com/dgoujard/uptimeWorker/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dgoujard/uptimeWorker/config"
)

func New(conf *config.Conf) (db *gorm.DB, err error) {
	var logLevel logger.LogLevel
	if conf.Debug {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=Europe/Paris",
		conf.Postgres.Server,
		conf.Postgres.User,
		conf.Postgres.Password,
		conf.Postgres.Database,
		conf.Postgres.Port)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil{
		return
	}
	// Migrate the schema
	err = db.AutoMigrate(&model.User{}, &model.NotificationGroup{},&model.SiteGroup{}, &model.LogType{}, &model.Site{}, &model.Log{})
	return
}