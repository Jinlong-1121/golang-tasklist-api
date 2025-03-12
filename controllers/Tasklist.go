package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	helper "go-todolist/helpers"
	"go-todolist/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetDepartemen godoc
//
//	@Router			/Tasklist/GetDepartemen [Get]
func (repository *InitRepo) GetDepartemen(c *gin.Context) {
	var departemen []models.DeptList
	// if err := c.ShouldBindJSON(&departemen); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	helper.MasterQuery = models.Query_MasterDept
	errs := helper.MasterExec_Get(repository.DbPg, &departemen)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  departemen,
	})

}

// GetTaskID godoc
//
//	@Router			/Tasklist/GetTaskID [Get]
func (repository *InitRepo) GetTaskID(c *gin.Context) {
	var Value []models.ValueGetTaskID
	var Parameter models.ParamGetTaskId

	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	helper.MasterQuery = models.Query_GettingTaskID + "'" + Parameter.Comment_id + "'"
	//fmt.Print(helper.MasterQuery)
	errs := helper.MasterExec_Get(repository.DbPg, &Value)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Value,
	})

}

// Category godoc
//
//	@Router			/Tasklist/GetCategory [Get]
func (repository *InitRepo) GetCategory(c *gin.Context) {
	var Topic []models.CategoryList
	var Parameter models.CategoryParam
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.GenerateValue_Category(Parameter.Param)
	helper.MasterQuery = models.GetCateoryList
	errs := helper.MasterExec_Get(repository.DbPg, &Topic)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Topic,
	})
}

// @Router			/Tasklist/GetUserid [Get]
func (repository *InitRepo) GetUserid(c *gin.Context) {
	var Value []models.ValueGettingUserid
	var Parameter models.ParamUserid
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if Parameter.Param == "GetUserid" {
		helper.MasterQuery = models.Query_GettingUserid + "'" + Parameter.Pin + "'"
	} else if Parameter.Param == "GetUserName" {
		helper.MasterQuery = models.Query_GettingUserName + "'" + Parameter.Pin + "'"
	}
	errs := helper.MasterExec_Get(repository.DbMy, &Value)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "errorrr"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Value,
	})
}

// @Router			/Tasklist/GetListtComments [Get]
func (repository *InitRepo) GetListtComments(c *gin.Context) {
	var Value []models.GetCommentList
	var Parameter models.ParamComments
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.GenerateValue_Comments(Parameter.Task_ID)
	helper.MasterQuery = models.QueryGetListComments
	errs := helper.MasterExec_Get(repository.DbPg, &Value)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Value,
	})
}

// ListData godoc
//
//	@Router			/Tasklist/GetListData [Get]
func (repository *InitRepo) GetListData(c *gin.Context) {
	var Parameter models.ListDataParams
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var Output interface{}

	switch Parameter.Param {
	case "GetDataHeaderTaskList":
		Output = []models.ListDataHeader{}
	case "GetDataDetailTaskList":
		Output = []models.ListDataDetail{}
	case "SetDataSummaryTaskList":
		Output = []models.ListDataSummary{}
	case "GetDataAssignTo", "GetDataAssignToALL":
		Output = []models.ListDataAssignTo{}
	case "ValidateUserLevel":
		Output = []models.ListDataValidateUserLevel{}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter"})
		return
	}

	models.GenerateValue_ListData(Parameter.Param, Parameter.Userid, Parameter.TaskID)
	helper.MasterQuery = models.QueryGetListData
	errs := helper.MasterExec_Get(repository.DbPg, &Output)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}

	// **Ensure the JSON response always includes expected columns**
	switch Parameter.Param {
	case "ValidateUserLevel":
		if data, ok := Output.([]models.ListDataValidateUserLevel); !ok || len(data) == 0 {
			Output = []models.ListDataValidateUserLevel{{Direct_Spv_No: ""}}
		}
	case "GetDataHeaderTaskList":
		if data, ok := Output.([]models.ListDataHeader); !ok || len(data) == 0 {
			Output = []models.ListDataHeader{{}} // Ensure empty struct in slice
		}
	case "GetDataDetailTaskList":
		if data, ok := Output.([]models.ListDataDetail); !ok || len(data) == 0 {
			Output = []models.ListDataDetail{{}}
		}
	case "SetDataSummaryTaskList":
		if data, ok := Output.([]models.ListDataSummary); !ok || len(data) == 0 {
			Output = []models.ListDataSummary{{}}
		}
	case "GetDataAssignTo", "GetDataAssignToALL":
		if data, ok := Output.([]models.ListDataAssignTo); !ok || len(data) == 0 {
			Output = []models.ListDataAssignTo{{}}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Output,
	})
}

// GetHeaderListData godoc
//
//	@Router			/Tasklist/GetHeaderListData [Get]
func (repository *InitRepo) GetHeaderListData(c *gin.Context) {
	var Parameter models.ListDataParams
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var Output interface{}
	if Parameter.Param == "GetDataHeaderTaskList" {
		Output = make([]models.ListDataHeader, 0)
	} else if Parameter.Param == "GetDataDetailTaskList" {
		Output = make([]models.ListDataDetail, 0)
	} else if Parameter.Param == "SetDataSummaryTaskList" {
		Output = make([]models.ListDataSummary, 0)
	} else if Parameter.Param == "GetDataAssignTo" {
		Output = make([]models.ListDataAssignTo, 0)
	} else if Parameter.Param == "ValidateUserLevel" {
		Output = make([]models.ListDataValidateUserLevel, 0)
	}
	models.GenerateValue_ListData(Parameter.Param, Parameter.Userid, Parameter.TaskID)
	helper.MasterQuery = models.QueryGetListData
	errs := helper.MasterExec_Get(repository.DbPg, &Output)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Output,
	})
}

// FetchData_Assign_To godoc
//
//	@Router			/Tasklist/FetchData_Assign_To [Get]
func (repository *InitRepo) FetchData_Assign_To(c *gin.Context) {

	apiURL := "http://192.168.10.23:6063/api/v1/data-assign-to/P0124006"
	response, err := http.Get(apiURL)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"error": true,
			"data":  err.Error(),
		})
		return
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"error": true,
			"data":  "Failed to read response body: " + err.Error(),
		})
		return
	}
	// Parse the JSON response
	var parsedResponse models.Response
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Create a table writer
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)

	// Print table header
	fmt.Fprintln(writer, "EMPLOYEE NAME\tEMPLOYEE NUMBER")

	// Print rows of data
	for _, record := range parsedResponse.Data {
		fmt.Fprintf(writer, "%s\t%s\n", record.Emp_Name, record.Emp_Name)
	}

	// Flush the table to display it
	writer.Flush()

	// Convert the parsed response back to JSON
	reshapedJSON, err := json.MarshalIndent(parsedResponse, "", "  ")
	if err != nil {
		fmt.Println("Error converting back to JSON:", err)
		return
	}

	// Print the reshaped JSON
	fmt.Println("\nConverted Back to JSON:")
	fmt.Println(string(reshapedJSON))

	c.JSON(http.StatusOK, reshapedJSON)
}

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	// remember close file
	defer f.Close()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

// @Param file body models.InsertComments true "Inserting Comments"
// @Router /Tasklist/InsertingComment [post]
func (repository *InitRepo) InsertingComment(c *gin.Context) {
	var AddingValue models.InsertComments
	var Try []models.CategoryList
	if err := c.ShouldBindJSON(&AddingValue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}
	if AddingValue.Comments == "TESTING" {
		content, err := readPdf(AddingValue.File_Path) // Read local pdf file
		if err != nil {
			panic(err)
		}
		fmt.Println(content)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded", "Content": content})
	} else {
		if AddingValue.File_Path == "" || len(AddingValue.File_Path) < 1 {
			AddingValue.File_Path = ""
			helper.MasterQuery = models.Query_InsertingComments + "('" + AddingValue.Task_ID + "','" + AddingValue.Comments + "','" + AddingValue.Emp_ID + "','" + AddingValue.File_Path + "','" + AddingValue.Content_Name + "','')"
			errs := helper.MasterExec_Get(repository.DbPg, &Try)
			if errs != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})

		} else {
			ObjID, err := helper.InsertPDFToMongoDB(AddingValue.File_Path)
			if err != nil {
				log.Printf("Failed to insert PDF into MongoDB: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload the PDF"})
				return
			}
			ObjectID := ObjID
			content, err := readPdf(AddingValue.File_Path) // Read local pdf file
			if err != nil {
				log.Printf("Failed to insert PDF into MongoDB: %v", err)
			}
			fmt.Println(content)
			c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})
			fmt.Println("PDF successfully inserted into MongoDB.")
			helper.MasterQuery = ""
			helper.MasterQuery = models.Query_InsertingComments + "('" + AddingValue.Task_ID + "','" + AddingValue.Comments + "','" + AddingValue.Emp_ID + "','" + ObjectID.Hex() + "','" + AddingValue.Content_Name + "')"
			errs := helper.MasterExec_Get(repository.DbPg, &Try)
			if errs != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})
		}
	}

}

// @Summary SendingNotifDone
// @Param file body models.WaitingToCloseEmail true "Inserting Task Manual"
// @Success 200 {object} map[string]string "Successfully uploaded"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /Tasklist/SendingNotifDone [post]
func (repository *InitRepo) SendingNotifDone(c *gin.Context) {
	var AddingValue models.WaitingToCloseEmail
	if err := c.ShouldBindJSON(&AddingValue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}
	var Userassignto []models.FetchUsernameAssign
	var UserReporter []models.FetchUsernameReporter
	var Taskidftch []models.FetchTaskID
	var Mailto []models.Mailto

	var username = ""
	helper.MasterQuery = `select "emp_no","emp_name" from public."dynamic_group" where "emp_no" =` + " '" + AddingValue.Assign_To + "' "
	errs_1 := helper.MasterExec_Get(repository.DbPg, &Userassignto)
	if errs_1 != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_1})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Userassignto,
	})
	username = Userassignto[0].Emp_Name

	var username_reporter = ""
	helper.MasterQuery = `select "emp_no","emp_name" from public."dynamic_group" where "emp_no" =` + " '" + AddingValue.Addwho + "' "
	errs_2 := helper.MasterExec_Get(repository.DbPg, &UserReporter)
	if errs_2 != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_2})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  UserReporter,
	})
	username_reporter = UserReporter[0].Emp_Name

	var Taskid = ""
	helper.MasterQuery = `select "task_id" from public."task_header" where "reporter" = ` + "'" + AddingValue.Addwho + "'" + ` order by "task_id" desc limit 1`
	errs_3 := helper.MasterExec_Get(repository.DbPg, &Taskidftch)
	if errs_3 != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_3})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Taskidftch,
	})
	Taskid = Taskidftch[0].Task_ID

	var SendMailto = ""
	helper.MasterQuery = `Select Email from users where number_officer =` + "'" + AddingValue.Addwho + "' "
	errs_4 := helper.MasterExec_Get(repository.DbMy, &Mailto)
	if errs_4 != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_4})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Mailto,
	})
	SendMailto = Mailto[0].Email
	enddate, err := time.Parse("2006-01-02", AddingValue.End_Date)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return
	}
	var clickdbtn = "<a style='background-color: rgb(255, 198, 39); color: white; padding: 15px 32px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; border-radius: 8px;' href='http://192.168.4.250/sipam/#/tasklist?Taskid=" + Taskid + "&Update=Close'>Close Your Task Here</a>"

	emailData := map[string]interface{}{
		"email_from":     "SiPAM Notifications (No-Reply)",
		"email_to":       SendMailto,                    // Jika lebih satu email kasih tnada koma (,)
		"email_cc":       "",                            // Jika lebih satu email kasih tnada koma (,)
		"email_template": "Email_Waiting_To_Close.html", // Sesuai dengan nama file HTML
		"email_subject":  "Notification",                // Subject Email bebas
		"email_body":     "",
		"param1":         username_reporter,
		"param2":         AddingValue.Subject,
		"param3":         "Done",
		"param4":         username,
		"param5":         enddate,
		"param6":         clickdbtn,
		"param7":         "",
		"param8":         "",
		"param9":         "",
		"param10":        "",
		"email_category": "Notification", // Email Catefory bebas
	}
	jsonData, err := json.Marshal(emailData)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"error": true,
			"data":  "Failed to marshal email data",
		})
		return
	}
	apiURL := "http://192.168.10.203:6069/api/v1/create_email_sender"
	response, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"error": true,
			"data":  err.Error(),
		})
		return
	}
	defer response.Body.Close()

	// Respond with success
	c.PureJSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  "Emails sent successfully",
	})
	c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})

}

// @Summary Inserting Subtask
// @Param file body models.InsertingTaskManual true "Inserting Task Manual"
// @Success 200 {object} map[string]string "Successfully uploaded"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /Tasklist/InsertingSubtask [post]
func (repository *InitRepo) InsertingSubtask(c *gin.Context) {

	var AddingValue models.InsertingTaskManual
	var Userassignto []models.FetchUsernameAssign
	var UserReporter []models.FetchUsernameReporter
	var Taskidftch []models.FetchTaskID
	var Mailto []models.Mailto
	if err := c.ShouldBindJSON(&AddingValue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	if strings.Contains(AddingValue.Assign_To, "GROUP") {
		var remainder_date string
		enddate, err := time.Parse("2006-01-02", AddingValue.End_Date)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}
		remainderDays, err := strconv.Atoi(AddingValue.Remainder_Date)
		if err != nil {
			fmt.Println("Error converting Remainder_Date to integer:", err)
			return
		}
		remainderDate := enddate.AddDate(0, 0, -remainderDays)
		remainder_date = remainderDate.Format("2006-01-02")
		helper.MasterQuery = models.Query_InsertSubtask + "('" + AddingValue.Departemen + "', '" + AddingValue.Topic + "', '" + AddingValue.Assign_To + "', '" + AddingValue.Priority + "','" + AddingValue.Subject + "', '" + AddingValue.Task_Name + "', '" + AddingValue.Start_Date + "', '" + AddingValue.End_Date + "', '" + AddingValue.Addwho + "','" + remainder_date + "','" + AddingValue.Task_id_parent_of + "')"
		errs := helper.MasterExec_Get(repository.DbPg, &AddingValue)
		if errs != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
			return
		}

	} else {
		var remainder_date string
		enddate, err := time.Parse("2006-01-02", AddingValue.End_Date)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}
		remainderDays, err := strconv.Atoi(AddingValue.Remainder_Date)
		if err != nil {
			fmt.Println("Error converting Remainder_Date to integer:", err)
			return
		}
		var username = ""
		helper.MasterQuery = `select "emp_no","emp_name" from public."dynamic_group" where "emp_no" =` + " '" + AddingValue.Assign_To + "' "
		errs_1 := helper.MasterExec_Get(repository.DbPg, &Userassignto)
		if errs_1 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_1})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  Userassignto,
		})
		username = Userassignto[0].Emp_Name

		var username_reporter = ""
		helper.MasterQuery = `select "emp_no","emp_name" from public."dynamic_group" where "emp_no" =` + " '" + AddingValue.Addwho + "' "
		errs_2 := helper.MasterExec_Get(repository.DbPg, &UserReporter)
		if errs_2 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_2})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  UserReporter,
		})
		username_reporter = UserReporter[0].Emp_Name
		var CurentDate = time.Now().Format("2006-01-02:15:04")

		remainderDate := enddate.AddDate(0, 0, -remainderDays)
		remainder_date = remainderDate.Format("2006-01-02")
		helper.MasterQuery = models.Query_InsertSubtask + "('" + AddingValue.Departemen + "', '" + AddingValue.Topic + "', '" + AddingValue.Assign_To + "', '" + AddingValue.Priority + "','" + AddingValue.Subject + "', '" + AddingValue.Task_Name + "', '" + AddingValue.Start_Date + "', '" + AddingValue.End_Date + "', '" + AddingValue.Addwho + "','" + remainder_date + "','" + AddingValue.Task_id_parent_of + "')"
		errs := helper.MasterExec_Get(repository.DbPg, &AddingValue)
		if errs != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
			return
		}

		var Taskid = ""
		helper.MasterQuery = `select "task_id" from public."task_header" where "reporter" = ` + "'" + AddingValue.Addwho + "'" + ` order by "task_id" desc limit 1`
		errs_3 := helper.MasterExec_Get(repository.DbPg, &Taskidftch)
		if errs_3 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_3})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  Taskidftch,
		})
		Taskid = Taskidftch[0].Task_ID

		var SendMailto = ""
		helper.MasterQuery = `Select Email from users where number_officer =` + "'" + AddingValue.Assign_To + "' "
		errs_4 := helper.MasterExec_Get(repository.DbMy, &Mailto)
		if errs_4 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_4})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  Mailto,
		})
		SendMailto = Mailto[0].Email

		var clickdbtn = "<a style='background-color: rgb(255, 198, 39); color: white; padding: 15px 32px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; border-radius: 8px;' href='http://192.168.4.250/sipam/#/tasklist?Taskid=" + Taskid + "'>Show Your Task Here</a>"

		emailData := map[string]interface{}{
			"email_from":     "SiPAM Notifications (No-Reply)",
			"email_to":       SendMailto,                    // Jika lebih satu email kasih tnada koma (,)
			"email_cc":       "",                            // Jika lebih satu email kasih tnada koma (,)
			"email_template": "Notifications_New_Task.html", // Sesuai dengan nama file HTML
			"email_subject":  "Notification",                // Subject Email bebas
			"email_body":     "",
			"param1":         username,
			"param2":         CurentDate,
			"param3":         AddingValue.Subject,
			"param4":         AddingValue.Remainder_Date,
			"param5":         username_reporter,
			"param6":         clickdbtn,
			"param7":         "",
			"param8":         "",
			"param9":         "",
			"param10":        "",
			"email_category": "Notification", // Email Catefory bebas
		}
		jsonData, err := json.Marshal(emailData)
		if err != nil {
			c.PureJSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": true,
				"data":  "Failed to marshal email data",
			})
			return
		}
		apiURL := "http://192.168.10.203:6069/api/v1/create_email_sender"
		response, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			c.PureJSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": true,
				"data":  err.Error(),
			})
			return
		}
		defer response.Body.Close()

		// Respond with success
		c.PureJSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  "Emails sent successfully",
		})
		c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})
	}
}

// @Summary Inserting Task Manual
// @Description Upload a file to the specified bucket using the file path and file name.
// @Accept json
// @Produce json
// @Param file body models.InsertingTaskManual true "Inserting Task Manual"
// @Success 200 {object} map[string]string "Successfully uploaded"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /Tasklist/InsertingTaskManual [post]
func (repository *InitRepo) InsertingTaskManual(c *gin.Context) {

	var AddingValue models.InsertingTaskManual
	var Userassignto []models.FetchUsernameAssign
	var UserReporter []models.FetchUsernameReporter
	var Taskidftch []models.FetchTaskID
	var Mailto []models.Mailto
	if err := c.ShouldBindJSON(&AddingValue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	if strings.Contains(AddingValue.Assign_To, "GROUP") {
		var remainder_date string
		enddate, err := time.Parse("2006-01-02", AddingValue.End_Date)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}
		remainderDays, err := strconv.Atoi(AddingValue.Remainder_Date)
		if err != nil {
			fmt.Println("Error converting Remainder_Date to integer:", err)
			return
		}
		remainderDate := enddate.AddDate(0, 0, -remainderDays)
		remainder_date = remainderDate.Format("2006-01-02")
		helper.MasterQuery = models.Query_InsertTaskManual + "('" + AddingValue.Departemen + "', '" + AddingValue.Topic + "', '" + AddingValue.Assign_To + "', '" + AddingValue.Priority + "','" + AddingValue.Subject + "', '" + AddingValue.Task_Name + "', '" + AddingValue.Start_Date + "', '" + AddingValue.End_Date + "', '" + AddingValue.Addwho + "','" + remainder_date + "')"
		errs := helper.MasterExec_Get(repository.DbPg, &AddingValue)
		if errs != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
			return
		}

	} else {
		var remainder_date string
		enddate, err := time.Parse("2006-01-02", AddingValue.End_Date)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}
		remainderDays, err := strconv.Atoi(AddingValue.Remainder_Date)
		if err != nil {
			fmt.Println("Error converting Remainder_Date to integer:", err)
			return
		}
		var username = ""
		helper.MasterQuery = `select "emp_no","emp_name" from public."dynamic_group" where "emp_no" =` + " '" + AddingValue.Assign_To + "' "
		errs_1 := helper.MasterExec_Get(repository.DbPg, &Userassignto)
		if errs_1 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_1})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  Userassignto,
		})
		username = Userassignto[0].Emp_Name

		var username_reporter = ""
		helper.MasterQuery = `select "emp_no","emp_name" from public."dynamic_group" where "emp_no" =` + " '" + AddingValue.Addwho + "' "
		errs_2 := helper.MasterExec_Get(repository.DbPg, &UserReporter)
		if errs_2 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_2})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  UserReporter,
		})
		username_reporter = UserReporter[0].Emp_Name
		var CurentDate = time.Now().Format("2006-01-02:15:04")

		remainderDate := enddate.AddDate(0, 0, -remainderDays)
		remainder_date = remainderDate.Format("2006-01-02")
		helper.MasterQuery = models.Query_InsertTaskManual + "('" + AddingValue.Departemen + "', '" + AddingValue.Topic + "', '" + AddingValue.Assign_To + "', '" + AddingValue.Priority + "','" + AddingValue.Subject + "', '" + AddingValue.Task_Name + "', '" + AddingValue.Start_Date + "', '" + AddingValue.End_Date + "', '" + AddingValue.Addwho + "','" + remainder_date + "')"
		errs := helper.MasterExec_Get(repository.DbPg, &AddingValue)
		if errs != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
			return
		}

		var Taskid = ""
		helper.MasterQuery = `select "task_id" from public."task_header" where "reporter" = ` + "'" + AddingValue.Addwho + "'" + ` order by "task_id" desc limit 1`
		errs_3 := helper.MasterExec_Get(repository.DbPg, &Taskidftch)
		if errs_3 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_3})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  Taskidftch,
		})
		Taskid = Taskidftch[0].Task_ID

		var SendMailto = ""
		helper.MasterQuery = `Select Email from users where number_officer =` + "'" + AddingValue.Assign_To + "' "
		errs_4 := helper.MasterExec_Get(repository.DbMy, &Mailto)
		if errs_4 != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_4})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  Mailto,
		})
		SendMailto = Mailto[0].Email

		var clickdbtn = "<a style='background-color: rgb(255, 198, 39); color: white; padding: 15px 32px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; border-radius: 8px;' href='http://192.168.4.250/sipam/#/tasklist?Taskid=" + Taskid + "'>Show Your Task Here</a>"

		emailData := map[string]interface{}{
			"email_from":     "SiPAM Notifications (No-Reply)",
			"email_to":       SendMailto,                    // Jika lebih satu email kasih tnada koma (,)
			"email_cc":       "",                            // Jika lebih satu email kasih tnada koma (,)
			"email_template": "Notifications_New_Task.html", // Sesuai dengan nama file HTML
			"email_subject":  "Notification",                // Subject Email bebas
			"email_body":     "",
			"param1":         username,
			"param2":         CurentDate,
			"param3":         AddingValue.Subject,
			"param4":         AddingValue.Remainder_Date,
			"param5":         username_reporter,
			"param6":         clickdbtn,
			"param7":         "",
			"param8":         "",
			"param9":         "",
			"param10":        "",
			"email_category": "Notification", // Email Catefory bebas
		}
		jsonData, err := json.Marshal(emailData)
		if err != nil {
			c.PureJSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": true,
				"data":  "Failed to marshal email data",
			})
			return
		}
		apiURL := "http://192.168.10.203:6069/api/v1/create_email_sender"
		response, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			c.PureJSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": true,
				"data":  err.Error(),
			})
			return
		}
		defer response.Body.Close()

		// Respond with success
		c.PureJSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  "Emails sent successfully",
		})
		c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})
	}

}

// @Summary Upload a file
// @Description Upload a file to the specified bucket using the file path and file name.
// @Accept json
// @Produce json
// @Param file body models.FileUpload.FilePath true "File Upload Info"
// @Success 200 {object} map[string]string "Successfully uploaded"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /Tasklist/UploadFile [post]
func (repository *InitRepo) UploadingFile(c *gin.Context) {
	bucketName := helper.GodotEnv("BucketName")
	var fileUpload models.FileUpload
	if err := c.ShouldBindJSON(&fileUpload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FilePath and FileName are required"})
		return
	}
	FileName := fileUpload.FileName
	FilePath := fileUpload.FilePath
	err := helper.UploadFile(bucketName, FilePath, FileName)
	if err != nil {
		fmt.Println("Error uploading file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})
}

// @Summary Upload a file
// @Accept json
// @Produce json
// @Param file body models.FileUpload.FilePath true "File Upload Info"
// @Success 200 {object} map[string]string "Successfully uploaded"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /Tasklist/UploadingToMongoDB_V1 [post]
func (repository *InitRepo) UploadingToMongoDB_V1(c *gin.Context) {
	var fileUpload models.FileUpload
	if err := c.ShouldBindJSON(&fileUpload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request, FilePath and FileName are required"})
		return
	}
	if fileUpload.FilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FilePath is required"})
		return
	}
	ObjID, err := helper.InsertPDFToMongoDB(fileUpload.FilePath)
	if err != nil {
		log.Printf("Failed to insert PDF into MongoDB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload the PDF"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded ,And This Your ID :" + ObjID.String()})
	fmt.Println("PDF successfully inserted into MongoDB.")

}

// @Summary Upload a file
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} map[string]string "Successfully uploaded"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /Tasklist/UploadingToMongoDB [post]
func (repository *InitRepo) UploadingToMongoDB(c *gin.Context) {
	// Parse the multipart form, with a maximum memory limit.
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	homeDir := os.TempDir()

	testingFolderPath := filepath.Join(homeDir, "DocumentTempSipam")

	tempFilePath := testingFolderPath + "\\" + file.Filename // Define a temporary path
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": tempFilePath})

	// ObjID, err := helper.InsertPDFToMongoDB(tempFilePath)
	// if err != nil {
	// 	log.Printf("Failed to insert PDF into MongoDB: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload the PDF"})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded, and this is your ID: " + ObjID.String()})
	// fmt.Println("PDF successfully inserted into MongoDB.")

	// // Optionally: clean up the temporary file after upload
	// os.Remove(tempFilePath)

}

// @Summary Upload a file
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Successfully uploaded"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /Tasklist/DownloadingToMongoDB [Get]
func (repository *InitRepo) DownloadingToMongoDB(c *gin.Context) {
	var Parameter models.ParamGetAttchment
	var output models.GettingFile
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	fmt.Printf("Could not get user home directory: %v\n", err)
	// 	return
	// }
	// testingFolderPath := filepath.Join(homeDir, "Downloads", "DocumentTempSipam")
	// if _, err := os.Stat(testingFolderPath); os.IsNotExist(err) {
	// 	err = os.MkdirAll(testingFolderPath, os.ModePerm) // Create the directory if it doesn't exist
	// 	if err != nil {
	// 		fmt.Printf("Could not create download directory: %v\n", err)
	// 		return
	// 	}
	// }
	// outputFilePath := filepath.Join(testingFolderPath, Parameter.FileName)
	var objectID = Parameter.ObjectID
	fileID, err := primitive.ObjectIDFromHex(objectID)
	if err != nil {
		fmt.Printf("Invalid ObjectID: %v\n", err)
		return
	}
	base64Data, filesize, err := helper.DownloadFileFromMongoDB(fileID)
	if err != nil {
		fmt.Printf("Error downloading file: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"error":   true,
			"message": "Error downloading file",
		})
	} else {
		fmt.Printf("File downloaded successfully to %s\n", "")
		output.Base64 = base64Data
		output.FilePath = strconv.Itoa(filesize)
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": false,
			"data":  output,
			// "base64": base64Data, // Include the Base64 string in the response
		})
	}

}

// @Summary Inserting Task Manual
// @Description Upload a file to the specified bucket using the file path and file name.
// @Accept json
// @Produce json
// @Param file body models.ValueUpdateingTask true "Updating Progress Task Value"
// @Success 200 {object} map[string]string "Successfully uploaded"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /Tasklist/UpdatingProgressTask [post]
func (repository *InitRepo) UpdatingProgressTask(c *gin.Context) {

	var AddingValue models.ValueUpdateingTask
	if err := c.ShouldBindJSON(&AddingValue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}
	models.GenerateValue_UpdateTask(AddingValue.Task_ID, AddingValue.ProgresValue)
	helper.MasterQuery = models.QueryUpdateTask
	errs := helper.MasterExec_Get(repository.DbPg, nil)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})
}

// Category godoc
//
//	@Router			/Tasklist/GetNotifTaskList [Get]
func (repository *InitRepo) GetNotifTaskList(c *gin.Context) {
	var Fetching []models.ColumnShowNotif
	var Parameter models.ParamShowNotif
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.GenerateValue_Notif(Parameter.UserID)
	helper.MasterQuery = models.Query_ShowNotif
	errs := helper.MasterExec_Get(repository.DbPg, &Fetching)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Fetching,
	})
}

// Category godoc
//
//	@Router			/Tasklist/GetUserNotifTaskList [Get]
func (repository *InitRepo) GetUserNotifTaskList(c *gin.Context) {
	var Fetching []models.ColumnShowUserNotif
	var Parameter models.ParamShowNotif
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.GenerateValue_Notif(Parameter.UserID)
	helper.MasterQuery = models.Query_ShowUserNotif
	errs := helper.MasterExec_Get(repository.DbPg, &Fetching)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Fetching,
	})

}

// Category godoc
//
//	@Router			/Tasklist/GetListUserAssignHistory [Get]
func (repository *InitRepo) GetListUserAssignHistory(c *gin.Context) {
	var Fetching []models.ColumnShowUserAssignHistory
	var Parameter models.ParamShowUserAssign_History
	if err := c.ShouldBindQuery(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.GenerateValue_UserAssignHistory(Parameter.Param)
	helper.MasterQuery = models.Query_userassignhistory
	errs := helper.MasterExec_Get(repository.DbPg, &Fetching)
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Fetching,
	})

}

// Category godoc
//
// @Param file body models.ParamClickedNotif true "File Upload Info"
//
//	@Router			/Tasklist/UpdateStatusClickednotif [Post]
func (repository *InitRepo) UpdateStatusClickednotif(c *gin.Context) {
	//var Fetching []models.ColumnShowUserNotif
	var Parameter models.ParamClickedNotif
	if err := c.ShouldBindJSON(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//models.GenerateValue_Notif(Parameter.UserID)
	helper.MasterQuery = models.Query_UpdateClickedNotif + "'" + Parameter.TaskID + "'"
	errs := helper.MasterExec_Get(repository.DbPg, "")
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  "Success",
	})
}

// @Param file body models.InsertUpdategroupAssignTOModels true "Inserting Data"
//
//	@Router			/Tasklist/InsertUpdategroupAssignTO [Post]
func (repository *InitRepo) InsertUpdategroupAssignTO(c *gin.Context) {
	//var Fetching []models.ColumnShowUserNotif
	var Parameter models.InsertUpdategroupAssignTOModels
	var Userassignto []models.FetchUsernameAssign
	var UserReporter []models.FetchUsernameReporter
	var Mailto []models.Mailto
	var Taskidftch []models.Getdetailtoreassign

	var CurentDate = time.Now().Format("2006-01-02:15:04")

	if err := c.ShouldBindJSON(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}
	helper.MasterQuery = models.Query_InsertUpdategroupAssignTO + "('" + Parameter.P_task_id + "','" + Parameter.P_user_assign_to + "','" + Parameter.P_group_assign + "','" + Parameter.P_assigner + "','" + Parameter.P_param + "')"
	errs := helper.MasterExec_Get(repository.DbPg, "")
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  "Success",
	})
	var username = ""
	helper.MasterQuery = `select "emp_no","emp_name" from public."dynamic_group" where "emp_no" =` + " '" + Parameter.P_user_assign_to + "' "
	errs_1 := helper.MasterExec_Get(repository.DbPg, &Userassignto)
	if errs_1 != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_1})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Userassignto,
	})
	username = Userassignto[0].Emp_Name

	var username_reporter = ""
	helper.MasterQuery = `select "emp_no","emp_name" from public."dynamic_group" where "emp_no" =` + " '" + Parameter.P_assigner + "' "
	errs_2 := helper.MasterExec_Get(repository.DbPg, &UserReporter)
	if errs_2 != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_2})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  UserReporter,
	})
	username_reporter = UserReporter[0].Emp_Name

	var SendMailto = ""
	helper.MasterQuery = `Select Email from users where number_officer =` + "'" + Parameter.P_user_assign_to + "' "
	errs_4 := helper.MasterExec_Get(repository.DbMy, &Mailto)
	if errs_4 != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_4})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Mailto,
	})
	SendMailto = Mailto[0].Email

	var Subject = ""
	helper.MasterQuery = `select "task_id","subject","estimated_time_done","created_at" as "created_date" from public."task_detail" where "task_id" = ` + "'" + Parameter.P_task_id + "'" + ` order by "task_id" desc limit 1`
	errs_3 := helper.MasterExec_Get(repository.DbPg, &Taskidftch)
	if errs_3 != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs_3})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  Taskidftch,
	})
	//const layout = "2006-01-02T15:04:05Z"
	//estimatedTime, err1 := time.Parse(layout, Taskidftch[0].Estimated_Time_Done)
	//createdTime, err2 := time.Parse(layout, Taskidftch[0].Created_Date)
	Subject = Taskidftch[0].Subject

	//if err1 != nil || err2 != nil {
	//	fmt.Println("Error parsing date:", err1, err2)
	//	return
	//}

	// Calculate the difference
	//Remainder_Date := estimatedTime.Sub(createdTime)

	var clickdbtn = "<a style='background-color: rgb(255, 198, 39); color: white; padding: 15px 32px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; border-radius: 8px;' href='http://192.168.4.250/sipam/#/tasklist?Taskid=" + Parameter.P_task_id + "'>Show Your Task Here</a>"

	emailData := map[string]interface{}{
		"email_from":     "SiPAM Notifications (No-Reply)",
		"email_to":       SendMailto,                    // Jika lebih satu email kasih tnada koma (,)
		"email_cc":       "",                            // Jika lebih satu email kasih tnada koma (,)
		"email_template": "Notifications_New_Task.html", // Sesuai dengan nama file HTML
		"email_subject":  "Notification",                // Subject Email bebas
		"email_body":     "",
		"param1":         username,
		"param2":         CurentDate,
		"param3":         Subject,
		"param4":         "",
		"param5":         username_reporter,
		"param6":         clickdbtn,
		"param7":         "",
		"param8":         "",
		"param9":         "",
		"param10":        "",
		"email_category": "Notification", // Email Catefory bebas
	}
	jsonData, err := json.Marshal(emailData)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"error": true,
			"data":  "Failed to marshal email data",
		})
		return
	}
	apiURL := "http://192.168.10.203:6069/api/v1/create_email_sender"
	response, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"error": true,
			"data":  err.Error(),
		})
		return
	}
	defer response.Body.Close()

	// Respond with success
	c.PureJSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  "Emails sent successfully ",
	})
	c.JSON(http.StatusOK, gin.H{"message": "Successfully uploaded"})
}

// @Param file body models.InsertSchedulerMasterTaskList true "Inserting Data"
//
//	@Router			/Tasklist/InsertSchedulerMasterTask [Post]
func (repository *InitRepo) InsertingSchedulerMasterTask(c *gin.Context) {
	//var Fetching []models.ColumnShowUserNotif
	var Parameter models.InsertSchedulerMasterTaskList
	if err := c.ShouldBindJSON(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}
	helper.MasterQuery = models.Query_InsertSchedulerMasterTaskList + "('" + Parameter.Topic_Code + "', '" + Parameter.Subject + "', '" + Parameter.Dept + "', '" + Parameter.Task_Name + "', '" + Parameter.Task_category + "', '" + Parameter.Generate_Every + "', '" + Parameter.Priority + "', '" + Parameter.Estimated_Time_Done + "', '" + Parameter.Assign_To + "', '" + Parameter.Remainder_Date + "','" + Parameter.Creator + "')"
	errs := helper.MasterExec_Get(repository.DbPg, "")
	if errs != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"error": false,
		"data":  "Success",
	})
}

// @Param file body models.CreateCategoryParam true "Inserting Data"
//
// @Router /Tasklist/CreateCategory [Post]
func (repository *InitRepo) CreateCategory(c *gin.Context) {
	var Parameter models.CreateCategoryParam

	if err := c.ShouldBindJSON(&Parameter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := repository.DbPg.Exec(`CALL public."SP_InsertingCategory"(?, ?)`, Parameter.Name, Parameter.Category).Error

	if err != nil {
		log.Println("Error calling stored procedure:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"error":   false,
		"message": "Category created successfully",
		"data":    Parameter,
	})
}
