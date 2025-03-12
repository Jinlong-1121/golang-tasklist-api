package controllers

import (
	"fmt"
	"go-todolist/configs"
	"go-todolist/models"

	"gorm.io/gorm"
)

type Controller struct{}

type InitRepo struct {
	DbPg *gorm.DB // For PostgreSQL
	DbMy *gorm.DB // For MySQL
}

// NewConnection initializes the database connections and returns an InitRepo instance
func NewConnection() *InitRepo {
	// Initialize both PostgreSQL and MySQL connections
	dbPg, err_1 := configs.InitDbPg()
	dbMy, err_2 := configs.InitDbMy()

	if err_1 != nil && err_2 != nil {
		fmt.Print("Connection Failed")
	} else {
		fmt.Print("Connection Success")
	}
	// Auto-migrate models for both databases
	dbPg.AutoMigrate(&models.DeptList{}) // For PostgreSQL
	dbMy.AutoMigrate(&models.DeptList{}) // For MySQL

	// Return the InitRepo with both database connections
	return &InitRepo{
		DbPg: dbPg,
		DbMy: dbMy,
	}
}

// NewController creates a new controller instance
func NewController() *Controller {
	return &Controller{}
}
