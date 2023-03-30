package models

import (
	"gorm.io/gorm"
	// ClusterModel "masterdb/models/clustermodel"
	// ConnectionProfileModel "masterdb/models/connectionprofilemodel"
	// ImportedDatabaseModel "masterdb/models/importeddatabasemodel"
	// TagModel "masterdb/models/tagmodel"
)

type User struct {
	gorm.Model
	Username           string
	Password           string
	Role               string // admin, consumer
	Host               string
	ConnectionProfiles []ConnectionProfile `gorm:"constraint:OnDelete:SET NULL;"`
	ImportedDatabases  []ImportedDatabase  `gorm:"constraint:OnDelete:SET NULL;"`
	Clusters           []Cluster           `gorm:"constraint:OnDelete:SET NULL;"`
	Tags               []Tag               `gorm:"constraint:OnDelete:SET NULL;"`
}
