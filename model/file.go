package model

import (
	"gorm.io/gorm"
	"time"
)

// File 文件模型，用于存储文件元数据
type File struct {
	ID          uint64         `gorm:"column:id;primary_key" json:"id"`
	FileName    string         `gorm:"column:file_name;not null" json:"file_name" crud:"allow_get,partial_update"`
	DisplayName string         `gorm:"column:display_name;not null" json:"display_name" crud:"allow_get,partial_update,required_on_create"`
	FileSize    int64          `gorm:"column:file_size;not null" json:"file_size" crud:"allow_get"`
	FileType    string         `gorm:"column:file_type;not null" json:"file_type" crud:"allow_get,partial_update"`
	FilePath    string         `gorm:"column:file_path;not null" json:"file_path" crud:"allow_get"`
	Uploader    uint64         `gorm:"column:uploader" json:"uploader" crud:"allow_get"`
	RelatedID   uint64         `gorm:"column:related_id" json:"related_id" crud:"allow_get,partial_update"`
	RelatedType string         `gorm:"column:related_type" json:"related_type" crud:"allow_get,partial_update"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at" crud:"allow_get"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at" crud:"allow_get"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (File) TableName() string {
	return "file"
}
