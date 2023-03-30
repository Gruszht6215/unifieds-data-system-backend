package database

import (
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"masterdb/models"
	// ClusterModel "masterdb/models/clustermodel"
	// ColumnModel "masterdb/models/columnmodel"
	// ConnectionProfileModel "masterdb/models/connectionprofilemodel"
	// ImportedDatabaseModel "masterdb/models/importeddatabasemodel"
	// TableModel "masterdb/models/tablemodel"
	// TagModel "masterdb/models/tagmodel"
	// UserModel "masterdb/models/usermodel"
)

var Db *gorm.DB
var err error

func InitDB() {
	dsn := os.Getenv("MARIADB_DSN")
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("***** failed to connect mariadb database *****")
	}

	Db.AutoMigrate(
		&models.User{},
		&models.ConnectionProfile{},
		&models.ImportedDatabase{},
		&models.Table{},
		&models.Column{},
		&models.Cluster{},
		&models.Tag{},
	)

	// disable soft delete for all operations on this GORM instance
	// Db.Unscoped().Delete(&User{}, &ConnectionProfile{}, &ImportedDatabase{})

}
