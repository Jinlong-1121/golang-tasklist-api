package models

type Datamentahdijoinkesini struct {
	Emp_No   int    `json:"emp_no"`
	Emp_Name string `json:"emp_name"`
}
type Response struct {
	Code  int                      `json:"code"`
	Data  []Datamentahdijoinkesini `json:"data"`
	Error bool                     `json:"error"`
}
type datamentah struct {
	Kpi_Option          string `json:"kpi_option"  gorm:"type:varchar(100);"`
	Subject             string `json:"subject"  gorm:"type:varchar(100);"`
	Task_ID             string `json:"task_id" gorm:"primary_key"`
	Task_Code           string `json:"task_code"  gorm:"type:varchar(100)"`
	Assign_To           string `json:"assign_to"  gorm:"type:varchar(100)"`
	Departemen          string `json:"departemen"  gorm:"type:varchar(100)"`
	Topic               string `json:"topic"  gorm:"type:varchar(100)"`
	Task_Progress       string `json:"task_progress"  gorm:"type:varchar(100)"`
	Estimated_Time_Done string `json:"estimated_time_done"  gorm:"type:timestamp"`
	Created_Date        string `json:"created_date"  gorm:"type:timestamp"`
	Start_Date          string `json:"start_date"  gorm:"type:timestamp"`
	Progress_Date       string `json:"progress_date"  gorm:"type:timestamp"`
	Finish_Date         string `json:"finish_date"  gorm:"type:timestamp"`
	Reporter            string `json:"reporter"  gorm:"type:varchar(100)"`
	Color               string `json:"color" gorm:"type:varchar(100)"`
	Task_id_parent_of   string `json:"task_id_parent_of" gorm:"type:varchar(100)"`
}

// type ListDataHeader_hasildarijoinan struct {
// 	Kpi_Option          string `json:"kpi_option"  gorm:"type:varchar(100);"`
// 	Subject             string `json:"subject"  gorm:"type:varchar(100);"`
// 	Task_ID             string `json:"task_id" gorm:"primary_key"`
// 	Task_Code           string `json:"task_code"  gorm:"type:varchar(100)"`
// 	Assign_To           string `json:"assign_to"  gorm:"type:varchar(100)"`
// 	Emp_Name            string `json:"emp_name"  gorm:"type:varchar(200)"`
// 	Departemen          string `json:"departemen"  gorm:"type:varchar(100)"`
// 	Topic               string `json:"topic"  gorm:"type:varchar(100)"`
// 	Task_Progress       string `json:"task_progress"  gorm:"type:varchar(100)"`
// 	Estimated_Time_Done string `json:"estimated_time_done"  gorm:"type:timestamp"`
// 	Created_Date        string `json:"created_date"  gorm:"type:timestamp"`
// 	Start_Date          string `json:"start_date"  gorm:"type:timestamp"`
// 	Progress_Date       string `json:"progress_date"  gorm:"type:timestamp"`
// 	Finish_Date         string `json:"finish_date"  gorm:"type:timestamp"`
// 	Reporter            string `json:"reporter"  gorm:"type:varchar(100)"`
// 	Color               string `json:"color" gorm:"type:varchar(100)"`
// 	Task_id_parent_of   string `json:"task_id_parent_of" gorm:"type:varchar(100)"`
// }

var fetchtasklistheader = ""

//= `SELECT  * FROM  PUBLIC."fetch_data_task_list_header"` + "('','','')" + "";

func Gen(Param string, Userid string, TaskID string) {
	fetchtasklistheader = `SELECT  * FROM  PUBLIC."fetch_data_task_list_header"` + "('','','')" + ""
}
