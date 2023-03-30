package models

import (
	"gorm.io/gorm"
)

type ImportedDatabase struct {
	gorm.Model
	Name                string
	Dbms                string
	Status              string // active, pending
	UserID              uint
	Description         string
	ConnectionProfileID uint
	Tables              []Table `gorm:"constraint:OnDelete:SET NULL;"`
}
