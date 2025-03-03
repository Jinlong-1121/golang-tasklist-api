package models

type DeptList struct {
	ID   int64  `json:"id" gorm:"primary_key;"`
	Name string `json:"name"  gorm:"type:varchar(100);"`
}
