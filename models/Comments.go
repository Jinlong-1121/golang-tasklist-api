package models

type InsertComments struct {
	Task_ID      string   `json:"Task_ID" binding:"required"`
	Comments     string   `json:"Comments" binding:"required"`
	Emp_ID       string   `json:"Emp_ID" binding:"required"`
	Content_Name string   `json:"Content_Name"`
	File_Path    string   `json:"File_Path"`
	Tagging_User []string `json:"Tagging_User"`
}

type GetCommentList struct {
	Comment_ID   string `json:"Comment_ID"  gorm:"type:varchar(100);"`
	Emp_ID       string `json:"Emp_ID"  gorm:"type:varchar(100);"`
	Emp_NAME     string `json:"Emp_NAME"  gorm:"type:varchar(100);"`
	Comment_Date string `json:"Comment_Date"  gorm:"type:timestamp;"`
	Comments     string `json:"Comments"  gorm:"type:varchar(100);"`
	Content_Name string `json:"Content_Name"  gorm:"type:varchar(100);"`
	File_ID      string `json:"File_ID"  gorm:"type:varchar(100);"`
}
type Configuration struct {
	RemoveUnused     bool   // Whether to remove unused objects
	ImageQuality     int    // Quality of images (0-100)
	CompressImages   bool   // Whether to compress images
	EnableEncryption bool   // Whether to enable encryption
	Password         string // Password for encryption if enabled
}
type ParamComments struct {
	Task_ID string `json:"task_id" gorm:"text;"`
}
type ParamGetAttchment struct {
	ObjectID string `json:"objectid" gorm:"text;"`
	FileName string `json:"filename" gorm:"text;"`
}

type GettingFile struct {
	FilePath string `json:"filepath" gorm:"text;"`
	Base64   []byte `json:"base64" gorm:"byte;"`
}

var QueryGetListComments = ""
var TableReturnedComments = `AS t("Comment_ID" varchar,"Emp_ID" varchar,"Emp_NAME" varchar,"Comment_Date" timestamp,"Comments" Text,"Content_Name" Text,"File_ID" Text)`

func GenerateValue_Comments(Param string) {
	QueryGetListComments = "Select * from public.Get_List_Comments('" + Param + "')" + TableReturnedComments
}

type InsertDocument struct {
	DocumentType string `json:"DocumentType" binding:"required"`
	CreatedDate  string `json:"CreatedDate" binding:"required"`
	Status       string `json:"Status" binding:"required"`
	TaskID       string `json:"TaskID" binding:"required"`
	DocumentName string `json:"DocumentName"`
	FilePath     string `json:"FilePath"`
}
