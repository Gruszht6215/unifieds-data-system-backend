package models

import (
	"gorm.io/gorm"
)

type Table struct {
	gorm.Model
	Name               string
	Description        string
	ImportedDatabaseID uint
	Columns            []Column `gorm:"constraint:OnDelete:SET NULL;"`
}
