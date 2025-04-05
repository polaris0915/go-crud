package model

import (
	"gorm.io/gorm"
	"time"
)

// RelateType 文件业务类型模型，定义系统中文件的业务类型
type RelateType struct {
	ID        uint64         `gorm:"type:bigint;unsigned;column:id;primary_key" json:"id" crud:"allow_get"`
	Type      string         `gorm:"type:varchar(255);column:type;" json:"type" crud:"required_on_create,allow_get"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at" crud:"allow_get"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at" crud:"allow_get"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (r *RelateType) TableName() string {
	return "relate_type"
}

// File 文件模型，用于存储文件元数据
type File struct {
	ID           uint64         `gorm:"type:bigint;unsigned;column:id;primary_key;uniqueIndex:idx_id_file_path,priority:1" json:"id"`
	FileName     string         `gorm:"type:varchar(255);column:file_name;not null" json:"file_name" crud:"allow_get,partial_update"`
	DisplayName  string         `gorm:"type:varchar(255);column:display_name;not null" json:"display_name" crud:"allow_get,partial_update,required_on_create"`
	FileSize     uint64         `gorm:"type:bigint;unsigned;column:file_size;not null" json:"file_size" crud:"allow_get"`
	FileType     string         `gorm:"type:varchar(30);column:file_type;not null" json:"file_type" crud:"allow_get,partial_update"`
	FilePath     string         `gorm:"type:text;column:file_path;not null;uniqueIndex:idx_id_file_path,priority:2,length:100" json:"file_path" crud:"allow_get"`
	Uploader     uint64         `gorm:"type:bigint;unsigned;column:uploader" json:"uploader" crud:"allow_get"`
	RelateTypeID uint64         `gorm:"type:bigint;unsigned;column:relate_type_id" json:"relate_type_id" crud:"allow_get"` // 用户不允许修改
	RelateType   *RelateType    `gorm:"foreignKey:RelateTypeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"relate_type"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at" crud:"allow_get"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at" crud:"allow_get"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (f *File) TableName() string {
	return "file"
}
