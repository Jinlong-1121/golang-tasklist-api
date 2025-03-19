package models

const (
	Query_MasterDept                    = `SELECT id, name FROM public."Master_Dept"`
	Query_InsertingComments             = `Call public."SP_InsertingComments"`
	Query_InsertTaskManual              = `Call public."SP_InsertingNewManualTask"`
	Query_InsertSubtask                 = `Call public."SP_InsertingSubtask"`
	Query_InsertUpdategroupAssignTO     = `Call public."user_assign_group_procedure"`
	Query_GettingUserid                 = `Select number_officer,name from users where pin = `
	Query_GettingUserName               = `Select number_officer,name from users where number_officer = `
	Query_GettingTaskID                 = `select "task_id" from public."task_comments" where "comment_id" = `
	Query_UpdateClickedNotif            = `Update public."user_notification_list" set "notif_status" = 'Clicked' where "notif_value" = `
	Query_InsertingNotif                = `Call public."SP_InsertNotif"`
	Query_InsertSchedulerMasterTaskList = `Call public."Sp_InsertingSchedulerTask"`
)

//("topic_code" text, "subject" text, "dept" text, "task_code" text, "task_name" text, "task_category" text, "generate_every" text, "priority" text, "estimasted_time_done" text, "assign_to" text, "created_date" text)
