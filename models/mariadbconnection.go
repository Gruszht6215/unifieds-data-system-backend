package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

type MariadbConnection struct {
	connectionProfile ConnectionProfile
}

func NewMariadbConnection(connectionProfile ConnectionProfile) *MariadbConnection {
	return &MariadbConnection{connectionProfile}
}

func (mc *MariadbConnection) ConnectDb() (*gorm.DB, error) {
	password := mc.connectionProfile.GetDecryptPassword()
	dsn := mc.connectionProfile.Username + ":" + password + "@tcp(" + mc.connectionProfile.Host + ":" + mc.connectionProfile.Port + ")/" + mc.connectionProfile.DatabaseName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db, err
}

func (mc *MariadbConnection) SyncSchema(targetDB *gorm.DB) error {
	var err error
	tables, err := mc.getMariadbTableList(targetDB)
	if err != nil {
		return err
	}
	if mc.getMariadbDescribeTable(targetDB, tables) != nil {
		log.Println("######## Failed to describe Mariadb tables ########\n", err.Error())
		return err
	}
	return nil
}

func (mc *MariadbConnection) getMariadbTableList(targetDB *gorm.DB) ([]string, error) {
	var tables []string
	if result := targetDB.Raw("SHOW TABLES").Scan(&tables); result.Error != nil {
		log.Println("######## Failed to get mariadb tables from target database ########\n", result.Error.Error())
		return nil, result.Error
	}
	return tables, nil
}

func (mc *MariadbConnection) getMariadbDescribeTable(targetDB *gorm.DB, tables []string) error {
	srcDb, err := mc.connectionProfile.conenctSrcDb()
	if err != nil {
		return err
	}
	srcDb.Preload("ImportedDatabase").Find(&mc.connectionProfile)

	for _, table := range tables {
		log.Println("######## Describe table: " + table + " ########")
		type DescribeTable struct {
			Field      string
			Type       string
			IsNullable string
			Key        string
			Default    *string
			Extra      string
		}

		describeTables := []DescribeTable{}
		if result := targetDB.Raw("DESCRIBE `" + table + "`").Scan(&describeTables); result.Error != nil {
			log.Println("######## Failed to describe tables ########\n", result.Error.Error())
			return result.Error
		}

		tableObj := Table{
			Name:               table,
			ImportedDatabaseID: mc.connectionProfile.ImportedDatabase.ID,
		}
		if result := srcDb.Create(&tableObj); result.Error != nil {
			log.Println("######## Failed to create table ########\n", table, result.Error.Error())
			return result.Error
		}

		var column Column
		for _, describeTable := range describeTables {
			var defaultValue string
			if describeTable.Default == nil {
				defaultValue = "NULL"
			} else {
				defaultValue = *describeTable.Default
			}
			var isNull = false
			if describeTable.IsNullable == "YES" {
				isNull = true
			}

			column = Column{
				Name:         describeTable.Field,
				Datatype:     describeTable.Type,
				IsNullable:   isNull,
				DefaultValue: defaultValue,
				Key:          describeTable.Key,
				Extra:        describeTable.Extra,
				TableID:      tableObj.ID,
			}
			// log.Printf("Field: %s, Type: %s, Null: %s, Key: %s, Default: %s, Extra: %s\n", describeTable.Field, describeTable.Type, describeTable.IsNullable, describeTable.Key, defaultValue, describeTable.Extra)
			if result := srcDb.Create(&column); result.Error != nil {
				log.Println("######## Failed to create column ########\n", table, result.Error.Error())
				return result.Error
			}
		}
	}
	return nil
}
