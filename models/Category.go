package models

type CategoryList struct {
	Code string `json:"code" gorm:"varchar(6);"`
	Name string `json:"name"  gorm:"type:varchar(100);"`
}
type CategoryParam struct {
	Param string `json:"Param" gorm:"varchar(30);"`
}
type CreateCategoryParam struct {
	Name     string `json:"name" binding:"required"`
	Category string `json:"category" binding:"required"`
}

// TaskCategory represents the task_category table structure
type TaskCategory struct {
	Name string `json:"name" gorm:"type:varchar(100);"`
}

var GetCateoryList = ""
var GetTaskCategoryList = ""

func GenerateValue_Category(param string) {
	GetCateoryList = "SELECT * FROM public.get_category('" + param + "') AS t(code VARCHAR, name VARCHAR)"
}

func GenerateValue_TaskCategory() {
	GetTaskCategoryList = "SELECT name FROM public.task_category"
}
