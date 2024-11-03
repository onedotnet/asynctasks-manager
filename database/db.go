package database

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/onedotnet/asynctasks/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type JSONB []interface{}

func (j JSONB) Value() (interface{}, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &j)
}

func TimeNow() time.Time {
	return time.Now().UTC()
}

func conn() (*gorm.DB, error) {
	cfg := config.AppConfig
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSL)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	sqlDB, err := conn.DB()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//conn.Logger.LogMode(logger.Info)
	sqlDB.SetConnMaxIdleTime(time.Second * 5)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(80)

	return conn, nil
}

var db *gorm.DB

func DB() *gorm.DB {
	if db == nil {
		conn, err := conn()
		if err != nil {
			log.Fatal(err)
		}
		db = conn
	}
	return db
}
