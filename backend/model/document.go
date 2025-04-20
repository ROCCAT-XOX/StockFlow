// backend/model/document.go
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Document repräsentiert ein Dokument oder eine Datei im System
type Document struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	FileName    string             `bson:"fileName" json:"fileName"`
	FileType    string             `bson:"fileType" json:"fileType"`
	Description string             `bson:"description" json:"description"`
	Category    string             `bson:"category" json:"category"`
	FilePath    string             `bson:"filePath" json:"filePath"`
	FileSize    int64              `bson:"fileSize" json:"fileSize"`
	UploadDate  time.Time          `bson:"uploadDate" json:"uploadDate"`
	UploadedBy  primitive.ObjectID `bson:"uploadedBy,omitempty" json:"uploadedBy"`
}
