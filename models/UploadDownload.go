package models

type FileUpload struct {
	FileName string `json:"fileName" binding:"required"`
	FilePath string `json:"filePath" binding:"required"`
}
