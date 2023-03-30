package models

import (
	"gorm.io/gorm"
)

type Cluster struct {
	gorm.Model
	Name   string
	UserID uint
	Tags   []*Tag `gorm:"many2many:cluster_tags;"`
}
