package models

import (
	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	Name        string
	Description string
	Color       string
	Clusters   []*Cluster `gorm:"many2many:cluster_tags;"`
	UserID      uint
	Columns    []*Column `gorm:"many2many:column_tags;"`
}
