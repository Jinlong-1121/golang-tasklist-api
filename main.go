package main

import (
	"fmt"
	"go-todolist/controllers"
	"go-todolist/cors"
	"go-todolist/docs"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Todo List API
// @version 1.0
// @description This is a sample server for a Todo List API.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

const (
	YYYYMMDD = "2006-01-02"
)

type InputRequest struct {
	Tes string `json:"tes"`
}

func main() {
	docs.SwaggerInfo.BasePath = "/api/v1"
	r := setupRouter()
	if err := r.Run(":6060"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func setupRouter() *gin.Engine {
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		fmt.Printf("Failed to create logs directory: %v\n", err)
		os.Exit(1)
	}
	now := time.Now().UTC()
	timeName := now.Format(YYYYMMDD)
	logFile, err := os.Create("logs/" + timeName + ".log")
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		os.Exit(1)
	}
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(logFile)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("{\"client_ip\":\"%s\", \"access_time\": \"%s\", \"method\": \"%s\", \"endpoint\": \"%s\", \"status_code\": %d, \"latency\": \"%s\", \"user_agent\": \"%s\", \"error\": \"%s\"}\n",
			param.ClientIP,
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	// registerRoutes(v1)

	initrepo := controllers.NewConnection()

	v1 := r.Group("/api/v1")
	{
		Tasklist := v1.Group("/Tasklist")
		{
			Tasklist.GET("/GetDepartemen", initrepo.GetDepartemen)
			Tasklist.GET("/GetIncomingTask", initrepo.GetIncomingTask)
			Tasklist.GET("/GetCategory", initrepo.GetCategory)
			Tasklist.GET("/GetTaskCategory", initrepo.GetTaskCategory)
			Tasklist.GET("/GetListData", initrepo.GetListData)
			Tasklist.GET("/ValidateDocType", initrepo.ValidateDocType)
			Tasklist.POST("/UploadFile", initrepo.UploadingFile)
			Tasklist.POST("/InsertingDocumentUpload", initrepo.InsertingDocumentUpload)
			Tasklist.POST("/InsertingComment", initrepo.InsertingComment)
			Tasklist.POST("/InsertingTaskManual", initrepo.InsertingTaskManual)
			Tasklist.POST("/UploadingToMongoDB", initrepo.UploadingToMongoDB)
			Tasklist.GET("/DownloadingToMongoDB", initrepo.DownloadingToMongoDB)
			Tasklist.GET("/GetUserid", initrepo.GetUserid)
			Tasklist.GET("/GetListUserAssignHistory", initrepo.GetListUserAssignHistory)
			Tasklist.GET("/GetListtComments", initrepo.GetListtComments)
			Tasklist.POST("/UpdatingProgressTask", initrepo.UpdatingProgressTask)
			Tasklist.GET("/GetNotifTaskList", initrepo.GetNotifTaskList)
			Tasklist.GET("/GetUserNotifTaskList", initrepo.GetUserNotifTaskList)
			Tasklist.GET("/FetchData_Assign_To", initrepo.FetchData_Assign_To)
			Tasklist.POST("/UpdateStatusClickednotif", initrepo.UpdateStatusClickednotif)
			Tasklist.POST("/InsertUpdategroupAssignTO", initrepo.InsertUpdategroupAssignTO)
			Tasklist.POST("/CreateCategory", initrepo.CreateCategory)
			Tasklist.POST("/InsertingSchedulerMasterTask", initrepo.InsertingSchedulerMasterTask)
			Tasklist.POST("/InsertingSubtask", initrepo.InsertingSubtask)
			Tasklist.POST("/SendingNotifDone", initrepo.SendingNotifDone)
			Tasklist.GET("/GetTaskID", initrepo.GetTaskID)
			Tasklist.GET("/MasterTagging", initrepo.MasterTagging)

		}
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8086")
	return r
}

// func registerRoutes(v1 *gin.RouterGroup) {
// 	v1.GET("/ping.php", pingHandler)
// 	v1.POST("/ping", pingPostHandler)
// 	v1.GET("/Tasklist/Departemen", controllers.GetDepartemen)
// }

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

func pingPostHandler(c *gin.Context) {
	var input InputRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, input)
}
