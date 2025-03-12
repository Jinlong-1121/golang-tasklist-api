package models

type ListDataHeader struct {
	Kpi_Option          string `json:"kpi_option"  gorm:"type:varchar(100);"`
	Subject             string `json:"subject"  gorm:"type:varchar(100);"`
	Task_ID             string `json:"task_id" gorm:"primary_key"`
	Task_Code           string `json:"task_code"  gorm:"type:varchar(100)"`
	Assign_To           string `json:"assign_to"  gorm:"type:varchar(100)"`
	Emp_Name            string `json:"emp_name"  gorm:"type:varchar(200)"`
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
type Getdetailtoreassign struct {
	Subject             string `json:"subject"  gorm:"type:varchar(100);"`
	Task_ID             string `json:"task_id" gorm:"primary_key"`
	Estimated_Time_Done string `json:"estimated_time_done"  gorm:"type:timestamp"`
	Created_Date        string `json:"created_date"  gorm:"type:timestamp"`
}
type ListDataDetail struct {
	Task_ID             string `json:"task_id" gorm:"primary_key;type:varchar(100);"`
	Task_Code           string `json:"task_code" gorm:"type:varchar(100);"`
	Departemen          string `json:"departemen" gorm:"type:varchar(100);"`
	Priority            string `json:"priority" gorm:"type:varchar(100);"`
	Topic               string `json:"topic" gorm:"type:varchar(100);"`
	Subject             string `json:"subject" gorm:"type:varchar(100);"`
	Task_Desc           string `json:"task_desc" gorm:"type:varchar(9999);"`
	Task_Progress       string `json:"task_progress" gorm:"type:varchar(100);"`
	Assign_To           string `json:"assign_to" gorm:"type:varchar(100);"`
	User_Assign_To      string `json:"user_assign_to" gorm:"type:varchar(100);"`
	Emp_Name            string `json:"emp_name" gorm:"type:varchar(100);"`
	Estimated_Time_Done string `json:"estimated_time_done" gorm:"type:timestamp;"`
	Start_Date          string `json:"start_date" gorm:"type:timestamp;"`
	Progress_Date       string `json:"progress_date" gorm:"type:timestamp;"`
	Finish_Date         string `json:"finish_date" gorm:"type:timestamp;"`
	Reporter            string `json:"reporter" gorm:"type:varchar(100);"`
}

type ListDataSummary struct {
	NEW         int64 `json:"new" gorm:"bigint;"`
	OPEN        int64 `json:"open" gorm:"bigint;"`
	IN_PROGRESS int64 `json:"in_progress"  gorm:"bigint;"`
	DONE        int64 `json:"done"  gorm:"bigint;"`
	HOLD        int64 `json:"hold"  gorm:"bigint;"`
	WARNING     int64 `json:"warning"  gorm:"bigint;"`
	OUTDATE     int64 `json:"outdate"  gorm:"bigint;"`
	TOTAL       int64 `json:"total"  gorm:"bigint;"`
}
type ListDataAssignTo struct {
	Emp_No   string `json:"emp_no" gorm:"type:varchar(100);"`
	Emp_Name string `json:"emp_name"  gorm:"type:varchar(100);"`
}

type ListDataValidateUserLevel struct {
	Direct_Spv_No   string `json:"direct_spv_no" gorm:"type:varchar(100);"`
	Direct_Spv_Name string `json:"direct_spv_name" gorm:"type:varchar(100);"`
	Group_Name      string `json:"group_name" gorm:"type:varchar(100);"`
}
type ListDataParams struct {
	Param  string `json:"Param" gorm:"varchar(50);"`
	Userid string `json:"userid" gorm:"varchar(30);"`
	TaskID string `json:"TaskID" gorm:"varchar(30);"`
}

var ReturnTableHeader = `("Kpi_Option" varchar,"Subject" varchar,"Task_ID" varchar,"Task_Code" varchar,"Assign_To" varchar,"emp_name" varchar,"Departemen" varchar,"Topic" varchar,"Task_Progress" varchar,"Estimated_Time_Done" timestamp,"Created_Date" timestamp,"Start_Date" timestamp,"Progress_Date" timestamp,"Finish_Date" timestamp,"Reporter" varchar,"Color" varchar,"Task_id_parent_of" varchar);`
var RetrunTableDetail = `("Task_ID" varchar,"Task_Code" varchar,"Departemen" varchar,"Priority" varchar,"Topic" varchar,"Subject" varchar,"Task_Desc" varchar,"Task_Progress" varchar,"Assign_To" varchar,"User_Assign_To" varchar,"Emp_Name" varchar,"Estimated_Time_Done" timestamp,"Start_Date" timestamp,"Progress_Date" timestamp,"Finish_Date" timestamp,"Reporter" varchar);`
var ReturnTableSummary = `("Emp_No" varchar,"Emp_Name" varchar,"NEW" bigint,"OPEN" bigint, "IN_PROGRESS" bigint, "DONE" bigint, "HOLD" bigint, "WARNING" bigint, "OUTDATE" bigint, "TOTAL" bigint);`
var RetrunTableAssignTo = `("emp_no" varchar,"emp_name" varchar);`
var RetrunTableValidateUserLevel = `("Direct_Spv_No" varchar,"Direct_Spv_Name" varchar,"Group_Name" varchar);`
var Tablereturn = ""

var QueryGetListData = ""

func GenerateValue_ListData(Param string, Userid string, TaskID string) {
	Tablereturn = ""
	if Param == "GetDataHeaderTaskList" {
		Tablereturn = ReturnTableHeader
	} else if Param == "GetDataDetailTaskList" {
		Tablereturn = RetrunTableDetail
	} else if Param == "SetDataSummaryTaskList" {
		Tablereturn = ReturnTableSummary
	} else if Param == "GetDataAssignTo" {
		Tablereturn = RetrunTableAssignTo
	} else if Param == "GetDataAssignToALL" {
		Tablereturn = RetrunTableAssignTo
	} else if Param == "ValidateUserLevel" {
		Tablereturn = RetrunTableValidateUserLevel
	} else if Param == "UpdateClickedNotif" {
		Tablereturn = RetrunTableAssignTo
	}
	QueryGetListData = "Select * from public.SP_New_Version_TaskList_Universal('" + Param + "','" + Userid + "','" + TaskID + "') AS " + Tablereturn
}

type InsertingTaskManual struct {
	Departemen        string `json:"departemen" `
	Topic             string `json:"topic"`
	Assign_To         string `json:"assign_to"`
	Priority          string `json:"priority"`
	Subject           string `json:"subject"`
	Task_Name         string `json:"task_name"`
	Start_Date        string `json:"start_date" `
	End_Date          string `json:"end_date" `
	Addwho            string `json:"addwho" `
	Remainder_Date    string `json:"remainder_date" `
	Task_id_parent_of string `json:"task_id_parent_of" `
}

type WaitingToCloseEmail struct {
	Task_id   string `json:"task_id" `
	Assign_To string `json:"assign_to"`
	Subject   string `json:"subject"`
	End_Date  string `json:"end_date" `
	Addwho    string `json:"addwho" `
}

type ParamUserid struct {
	Param string `json:"Param" gorm:"varchar(30);"`
	Pin   string `json:"pin" gorm:"varchar(30);"`
}

type ParamGetTaskId struct {
	Comment_id string `json:"comment_id" gorm:"varchar(30);"`
}
type ValueGettingUserid struct {
	Number_officer string `json:"number_officer" gorm:"varchar(30);"`
	Name           string `json:"name" gorm:"varchar(100);"`
}
type FetchUsernameAssign struct {
	Emp_No   string `json:"emp_no" gorm:"varchar(30);"`
	Emp_Name string `json:"emp_name" gorm:"varchar(100);"`
}
type FetchUsernameReporter struct {
	Emp_No   string `json:"emp_no" gorm:"varchar(30);"`
	Emp_Name string `json:"emp_name" gorm:"varchar(100);"`
}
type FetchTaskID struct {
	Task_ID string `json:"task_id" gorm:"varchar(12);"`
}
type Mailto struct {
	Email string `json:"email" gorm:"varchar(100);"`
}

var QueryUpdateTask = ""

type ValueUpdateingTask struct {
	Task_ID      string `json:"task_id" gorm:"varchar(30);"`
	ProgresValue string `json:"progresvalue" gorm:"varchar(30);"`
}
type ValueGetTaskID struct {
	Task_ID string `json:"task_id" gorm:"varchar(30);"`
}

func GenerateValue_UpdateTask(TaskID string, ProgresValue string) {
	QueryUpdateTask = `Call public."SP_Update_TaskProgress"` + "('" + TaskID + "', '" + ProgresValue + "')"
}

type ParamShowNotif struct {
	UserID string `json:"userid" gorm:"varchar(30);"`
}
type ParamClickedNotif struct {
	TaskID string `json:"taskid" gorm:"varchar(30);"`
}

type ColumnShowNotif struct {
	Current_Task int64 `json:"current_task" gorm:"bigint;"`
	Old_Task     int64 `json:"old_task" gorm:"bigint;"`
	New_Task     int64 `json:"new_task" gorm:"bigint;"`
}
type ColumnShowUserNotif struct {
	Officer_Number string `json:"officer_number" gorm:"varchar;"`
	Notif_Category string `json:"notif_category" gorm:"varchar;"`
	Notif_Value    string `json:"notif_value" gorm:"varchar;"`
	Notif_Status   string `json:"notif_status" gorm:"varchar;"`
	Created_at     string `json:"created_at" gorm:"timestamp;"`
	Subject        string `json:"subject" gorm:"varchar;"`
}

var tablereturnNotif = `t("Current_Task" bigint, "Old_Task" bigint, "New_Task" bigint)Limit 1;`
var tablereturnNotif_count = `t("Officer_Number" varchar, "Notif_Category" varchar, "Notif_Value" varchar, "Notif_Status" Varchar, "Created_at" timestamp, "Subject" varchar);`
var Query_ShowNotif = ""
var Query_ShowUserNotif = ""

func GenerateValue_Notif(UserID string) {
	Query_ShowNotif = "Select * from public.tasknotification('" + UserID + "') AS " + tablereturnNotif
	Query_ShowUserNotif = "Select * from public.User_Notification('" + UserID + "') AS " + tablereturnNotif_count
}

type InsertUpdategroupAssignTOModels struct {
	P_task_id        string `json:"p_task_id"`
	P_user_assign_to string `json:"p_user_assign_to"`
	P_group_assign   string `json:"p_group_assign"`
	P_assigner       string `json:"p_assigner"`
	P_param          string `json:"p_param"`
}

type InsertSchedulerMasterTaskList struct {
	Topic_Code          string `json:"topic_code"`
	Subject             string `json:"subject"`
	Dept                string `json:"dept"`
	Task_Name           string `json:"task_name"`
	Task_category       string `json:"task_category"`
	Generate_Every      string `json:"generate_every"`
	Priority            string `json:"priority"`
	Estimated_Time_Done string `json:"estimasted_time_done"`
	Assign_To           string `json:"assign_to"`
	Remainder_Date      string `json:"remainder_date"`
	Creator             string `json:"creator"`
}

//("" text, "" text, "" text, "" text, "" text, "" text, "" text, "" text, "" text, "" text, "" text)

type ParamShowUserAssign_History struct {
	Param string `json:"param" gorm:"varchar(30);"`
}

type ColumnShowUserAssignHistory struct {
	Assigner       string `json:"assigner" gorm:"varchar;"`
	Assigner_name  string `json:"assigner_name" gorm:"varchar;"`
	User_assign_to string `json:"user_assign_to" gorm:"varchar;"`
	Emp_name       string `json:"emp_name" gorm:"varchar;"`
	Start_date     string `json:"start_date" gorm:"timestamp;"`
	End_date       string `json:"end_date" gorm:"timestamp;"`
	Duration       string `json:"duration" gorm:"varchar;"`
	Status         string `json:"status" gorm:"varchar;"`
}

var tablereturnuserassignhistory = `t("assigner" varchar,"assigner_name" text,"user_assign_to" varchar,"emp_name" varchar,"start_date" TIMESTAMP,"end_date" TIMESTAMP,"duration" text,"status" text)`

var Query_userassignhistory = ""

func GenerateValue_UserAssignHistory(Param string) {
	Query_userassignhistory = `Select * from "public"."User_Assign_History"` + "('" + Param + "') AS " + tablereturnuserassignhistory
}
