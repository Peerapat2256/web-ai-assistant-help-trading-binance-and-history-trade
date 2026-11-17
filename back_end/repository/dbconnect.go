package repository

import (
	"backend/model/model"
	"crypto/tls"
	"log"
	"os"
	"strings"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func getEnvWithFallback(primary, fallback string) string {
	val := strings.TrimSpace(os.Getenv(primary))
	if val != "" {
		return val
	}
	return strings.TrimSpace(os.Getenv(fallback))
}

func NewDbConnection() (*gorm.DB, error) {
	if db == nil {
		viper.SetConfigName("config")
		viper.AddConfigPath("./system")
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}

		if err := godotenv.Load(); err != nil {
			_ = godotenv.Load("../.env")
		}

		dbHost := getEnvWithFallback("DB_HOST", "HOST")
		dbPort := getEnvWithFallback("DB_PORT", "PORT")
		dbUser := getEnvWithFallback("DB_USER", "USERNAME")
		dbPass := getEnvWithFallback("DB_PASSWORD", "PASSWORD")
		dbName := getEnvWithFallback("DB_NAME", "DATABASE")

		// ลงทะเบียน TLS สำหรับเชื่อมต่อ TiDB Cloud
		_ = mysqldriver.RegisterTLSConfig("tidb", &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: dbHost,
		})

		dsn := viper.GetString("mysql.dsn")
		dsn = strings.Replace(dsn, "DB_USER", dbUser, 1)
		dsn = strings.Replace(dsn, "DB_PASSWORD", dbPass, 1)
		dsn = strings.Replace(dsn, "DB_HOST", dbHost, 1)
		dsn = strings.Replace(dsn, "DB_PORT", dbPort, 1)
		dsn = strings.Replace(dsn, "DB_NAME", dbName, 1)

		database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		log.Println("เชื่อมต่อ TiDB Cloud สำเร็จ")

		if err := database.AutoMigrate(
			&model.Member{},
			&model.PortfolioSetting{},
			&model.TradeHistory{},
			&model.AiAnalysisLog{},
			&model.SheetConfig{},
		); err != nil {
			log.Printf("คำเตือน AutoMigrate: %v", err)
		} else {
			log.Println("ตรวจสอบและสร้างตารางฐานข้อมูลครบถ้วนแล้ว")
		}

		sqlDB, err := database.DB()
		if err == nil {
			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetMaxOpenConns(50)
			sqlDB.SetConnMaxLifetime(time.Hour)
		}
		db = database
	}
	return db, nil
}
