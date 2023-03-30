package models

import (
	"gorm.io/gorm"
)

type Column struct {
	gorm.Model
	Name         string
	Datatype     string
	IsNullable   bool
	Key          string
	DefaultValue string
	Extra        string
	Description  string
	TableID      uint
	Tags         []*Tag `gorm:"many2many:column_tags;"`
}

func (dt *Column) ToString() string {
	return dt.Name + " " + dt.Datatype + " " + dt.Key + " " + dt.Extra
}
