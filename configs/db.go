package configs

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	helper "go-todolist/helpers"
)

var (
	DB_USERNAME_PG = helper.GodotEnv("DB_USERNAME_PG")
	DB_PASSWORD_PG = helper.GodotEnv("DB_PASSWORD_PG")
	DB_NAME_PG     = helper.GodotEnv("DB_NAME_PG")
	DB_HOST_PG     = helper.GodotEnv("DB_HOST_PG")
	DB_PORT_PG     = helper.GodotEnv("DB_PORT_PG")

	DB_USERNAME_MY = helper.GodotEnv("DB_USERNAME_MY")
	DB_PASSWORD_MY = helper.GodotEnv("DB_PASSWORD_MY")
	DB_NAME_MY     = helper.GodotEnv("DB_NAME_MY")
	DB_HOST_MY     = helper.GodotEnv("DB_HOST_MY")
	DB_PORT_MY     = helper.GodotEnv("DB_PORT_MY")
)

func InitDbPg() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		DB_HOST_PG, DB_USERNAME_PG, DB_PASSWORD_PG, DB_NAME_PG, DB_PORT_PG)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Print("Error connecting to database 01 : error=", err)
		return nil, err
	} else {
		fmt.Println("Db_01 Connected")
	}

	return db, nil
}

func InitDbMy() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		DB_USERNAME_MY, DB_PASSWORD_MY, DB_HOST_MY, DB_PORT_MY, DB_NAME_MY)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Print("Error connecting to database 02 : error=", err)
		return nil, err
	} else {
		fmt.Println("Db_02 Connected")
	}

	return db, nil
}
