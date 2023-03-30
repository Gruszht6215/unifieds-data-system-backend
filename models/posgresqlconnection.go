package models

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type PostgresqlConnection struct {
	connectionProfile ConnectionProfile
}

func NewPostgresqlConnection(connectionProfile ConnectionProfile) *PostgresqlConnection {
	return &PostgresqlConnection{connectionProfile}
}

func (pc *PostgresqlConnection) ConnectDb() (*gorm.DB, error) {
	password := pc.connectionProfile.GetDecryptPassword()
	dsn := "host=" + pc.connectionProfile.Host + " port=" + pc.connectionProfile.Port + " user=" + pc.connectionProfile.Username + " dbname=" + pc.connectionProfile.DatabaseName + " password=" + password + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db, err
}

func (pc *PostgresqlConnection) SyncSchema(targetDB *gorm.DB) error {
	var err error
	tables, err := pc.getPostgresTableList(targetDB)
	if err != nil {
		return err
	}
	if pc.getPostgresDescribeTable(targetDB, tables) != nil {
		log.Println("######## Failed to describe Postgres tables ########\n", err.Error())
		return err
	}
	return nil
}

func (pc *PostgresqlConnection) getPostgresTableList(targetDB *gorm.DB) ([]string, error) {
	var tables []string
	sql := "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public'"
	if result := targetDB.Raw(sql).Scan(&tables); result.Error != nil {
		log.Println("######## Failed to get postgresql tables from target database ########\n", result.Error.Error())
		return nil, result.Error
	}
	return tables, nil
}

func (pc *PostgresqlConnection) getPostgresDescribeTable(targetDB *gorm.DB, tables []string) error {
	srcDb, err := pc.connectionProfile.conenctSrcDb()
	if err != nil {
		return err
	}
	srcDb.Preload("ImportedDatabase").Find(&pc.connectionProfile)

	for _, table := range tables {
		type DescribeTable struct {
			ColumnName    string
			DataType      string
			IsNullable    string
			ColumnDefault string
		}
		describeTables := []DescribeTable{}
		sqlRaw := "SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_name = '" + table + "';"
		if result := targetDB.Raw(sqlRaw).Scan(&describeTables); result.Error != nil {
			log.Println("######## Failed to describe tables ########\n", result.Error.Error())
			return result.Error
		}
		sqlRaw = "SELECT a.attname FROM pg_index i JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey) WHERE i.indrelid = '" + table + "'::regclass AND i.indisprimary;"
		var primaryKeys []string
		if result := targetDB.Raw(sqlRaw).Scan(&primaryKeys); result.Error != nil {
			log.Println("######## Failed to get tables primary key ########\n", result.Error.Error())
			return result.Error
		}

		tableObj := Table{
			Name:               table,
			ImportedDatabaseID: pc.connectionProfile.ImportedDatabase.ID,
		}
		if result := srcDb.Create(&tableObj); result.Error != nil {
			log.Println("######## Failed to create table ########\n", table, result.Error.Error())
			return result.Error
		}

		var column Column
		for _, describeTable := range describeTables {
			var isNull = false
			if describeTable.IsNullable == "YES" {
				isNull = true
			}

			column = Column{
				Name:         describeTable.ColumnName,
				Datatype:     describeTable.DataType,
				IsNullable:   isNull,
				DefaultValue: describeTable.ColumnDefault,
				TableID:      tableObj.ID,
			}
			for i, primaryKey := range primaryKeys {
				if primaryKey == describeTable.ColumnName {
					column.Key = "PRI"
					primaryKeys = pc.removeStringAtIndex(primaryKeys, i)
				}
			}

			if result := srcDb.Create(&column); result.Error != nil {
				log.Println("######## Failed to create database table ########\n", table, result.Error.Error())
				return result.Error
			}
		}
	}
	return nil
}

func (pc *PostgresqlConnection) removeStringAtIndex(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
