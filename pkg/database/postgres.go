package database

import (
	"fmt"
	"os"
	"time" // Don't forget this import

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectPostgres() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		panic("Failed to connect to database!")
	}

	// --- [NEW] PRODUCTION POOLING SETTINGS ---
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Maximum number of idle connections in the pool.
	sqlDB.SetMaxIdleConns(10)

	// Maximum number of open connections to the database.
	// 100 is safe for standard Postgres.
	sqlDB.SetMaxOpenConns(100)

	// Maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)
	// -----------------------------------------

	return db
}
