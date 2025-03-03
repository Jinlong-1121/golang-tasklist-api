package helper

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

var MasterQuery string

type MasterStruct struct{}

func MasterExec_Get(db *gorm.DB, masterStruct interface{}) (err error) {
	err = db.Raw(MasterQuery).Scan(masterStruct).Error
	if err != nil {
		return err
	}
	return nil
}
func MasterExec_Post(db *gorm.DB) (err error) {
	db.Exec("SET client_min_messages TO WARNING")
	err = db.Raw(MasterQuery).Commit().Error
	if err != nil {
		return err
	}
	return nil
}

func UploadFile(bucketName, filePath, key string) (err error) {
	// R2 credentials and endpoint
	accessKey := GodotEnv("accessKey")
	secretKey := GodotEnv("secretKey")
	endpoint := GodotEnv("endpoint")

	// Initialize an AWS session with the R2 endpoint
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("auto"),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true), // required for Cloudflare R2
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	svc := s3.New(sess)

	// Open the file to upload
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filePath, err)
	}
	defer file.Close()

	// Upload the file
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	fmt.Println("File successfully uploaded to R2")
	return nil
}

func InsertPDFToMongoDB_V1(Filepath string) (primitive.ObjectID, error) {
	uri := GodotEnv("Mongodb_Url")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return primitive.NilObjectID, err
	}
	defer client.Disconnect(ctx)
	database := client.Database(GodotEnv("DataBaseName"))
	bucket, err := gridfs.NewBucket(database, options.GridFSBucket().SetName("StoreDocTesting"))
	if err != nil {
		return primitive.NilObjectID, err
	}
	pdfFile, err := os.Open(Filepath)
	if err != nil {
		return primitive.NilObjectID, err
	}
	defer pdfFile.Close()
	fileName := filepath.Base(Filepath)
	uploadStream, err := bucket.OpenUploadStream(fileName)
	if err != nil {
		return primitive.NilObjectID, err
	}
	defer uploadStream.Close()
	_, err = io.Copy(uploadStream, pdfFile)
	if err != nil {
		return primitive.NilObjectID, err
	}

	fileID := uploadStream.FileID.(primitive.ObjectID)
	fmt.Printf("File uploaded successfully. File ID: %s\n", fileID.Hex())
	return fileID, nil
}

func InsertPDFToMongoDB(Filepath string) (primitive.ObjectID, error) {
	uri := GodotEnv("Mongodb_Url")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return primitive.NilObjectID, err
	}
	defer client.Disconnect(ctx)
	database := client.Database(GodotEnv("DataBaseName"))
	bucket, err := gridfs.NewBucket(database, options.GridFSBucket().SetName("StoreDoc"))
	if err != nil {
		return primitive.NilObjectID, err
	}
	pdfFile, err := os.Open(Filepath)
	if err != nil {
		return primitive.NilObjectID, err
	}
	defer pdfFile.Close()
	fileName := filepath.Base(Filepath)
	uploadStream, err := bucket.OpenUploadStream(fileName)
	if err != nil {
		return primitive.NilObjectID, err
	}
	defer uploadStream.Close()
	_, err = io.Copy(uploadStream, pdfFile)
	if err != nil {
		return primitive.NilObjectID, err
	}

	fileID := uploadStream.FileID.(primitive.ObjectID)
	fmt.Printf("File uploaded successfully. File ID: %s\n", fileID.Hex())
	return fileID, nil
}

func DownloadFileFromMongoDB(fileID primitive.ObjectID) ([]byte, int, error) {
	uri := GodotEnv("Mongodb_Url")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, 0, fmt.Errorf("could not connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	database := client.Database(GodotEnv("DataBaseName"))
	bucket, err := gridfs.NewBucket(database, options.GridFSBucket().SetName("StoreDoc"))
	if err != nil {
		return nil, 0, fmt.Errorf("could not create GridFS bucket: %w", err)
	}

	downloadStream, err := bucket.OpenDownloadStream(fileID)
	if err != nil {
		return nil, 0, fmt.Errorf("could not open download stream: %w", err)
	}
	defer downloadStream.Close()

	// Create a buffer to hold the file content
	fileContent, err := io.ReadAll(downloadStream)
	if err != nil {
		return nil, 0, fmt.Errorf("could not read file content: %w", err)
	}

	// Compress the PDF content
	compressedContent, err := compressPDF(fileContent)
	if err != nil {
		return nil, 0, fmt.Errorf("could not compress PDF: %w", err)
	}

	// Get the size of the compressed PDF file
	fileSize := len(compressedContent)

	fmt.Printf("File %s successfully downloaded from MongoDB. Compressed size: %d bytes\n", fileID.Hex(), fileSize)
	return compressedContent, fileSize, nil
}

// compressPDF compresses the PDF content using pdfcpu
func compressPDF(content []byte) ([]byte, error) {
	// Create a temporary file to store the uncompressed PDF
	tempInputFile := "temp_input.pdf"
	tempOutputFile := "temp_output.pdf"

	// Write the content to the temporary input file
	if err := os.WriteFile(tempInputFile, content, 0644); err != nil {
		return nil, fmt.Errorf("could not write temporary input file: %w", err)
	}
	defer os.Remove(tempInputFile) // Clean up

	// conf := &models.Configuration{
	// 	// Set the appropriate fields based on the library's documentation
	// 	RemoveUnused: true, // Example field, adjust according to your model's structure
	// 	ImageQuality: 75,   // Example field, adjust according to your model's structure
	// 	// Add other configuration options as needed
	// }
	// Compress the PDF
	if err := api.OptimizeFile(tempInputFile, tempOutputFile, nil); err != nil {
		return nil, fmt.Errorf("could not optimize PDF: %w", err)
	}
	defer os.Remove(tempOutputFile) // Clean up

	// Read the compressed content
	compressedContent, err := os.ReadFile(tempOutputFile)
	if err != nil {
		return nil, fmt.Errorf("could not read compressed output file: %w", err)
	}

	return compressedContent, nil
}
