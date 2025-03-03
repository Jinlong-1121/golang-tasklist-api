package controllers

import (
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
	dbPg, _ := configs.InitDbPg()
	dbMy, _ := configs.InitDbMy()

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
