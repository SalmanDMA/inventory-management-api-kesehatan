package configs

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	DB_NAME := os.Getenv("DB_NAME")
	DB_USERNAME := os.Getenv("DB_USERNAME")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_HOST := os.Getenv("DB_HOST")

	ENVIRONMENT := os.Getenv("ENVIRONMENT")

	SSLMODE := "disable"

	if ENVIRONMENT == "PRODUCTION" {
		SSLMODE = "require"
	}

	POSTGRESQL := "postgres://" + DB_USERNAME + ":" + DB_PASSWORD + "@" + DB_HOST + "/" + DB_NAME + "?sslmode=" + SSLMODE + "&TimeZone=Asia/Jakarta"

	dsn := POSTGRESQL
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}
	
	fmt.Println("Connection Opened to Database")
}