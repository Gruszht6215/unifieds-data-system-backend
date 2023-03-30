package models

import (
	"log"
	"os"
	"strings"
	// "unicode/utf8"

	Aes "masterdb/pkg/encryption"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DbConnecter interface {
	ConnectDb() (*gorm.DB, error)
	SyncSchema(*gorm.DB) error
}

type ConnectionProfile struct {
	gorm.Model
	Dbms             string // mysql, postgresql, mariadb
	ConnectionName   string
	Host             string
	Port             string
	DatabaseName     string
	Username         string
	Password         string
	UserID           uint
	ImportedDatabase ImportedDatabase `gorm:"constraint:OnDelete:CASCADE;"`
}

func (cp *ConnectionProfile) ToString() string {
	return cp.ConnectionName + " " + cp.Host + " " + cp.Port + " " + cp.DatabaseName + " " + cp.Username + " " + cp.Password
}

func HideSensitiveData(cp *ConnectionProfile) {
	// cp.Password = strings.Repeat("*", utf8.RuneCountInString(cp.Password))
	cp.UserID = 0
}

func (cp *ConnectionProfile) EncryptPassword() {
	cryptoText := Aes.Encrypt(cp.Password)
	cp.Password = cryptoText
}

func (cp ConnectionProfile) GetDecryptPassword() string {
	plainText := Aes.Decrypt(cp.Password)
	return plainText
}

func (cp *ConnectionProfile) GetDbms() string {
	return cp.Dbms
}

func (cp *ConnectionProfile) ConnectTargetDb() (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	var conn DbConnecter

	switch strings.ToLower(cp.Dbms) {
	case "mysql":
		conn = NewMysqlConnection(*cp)
	case "mariadb":
		conn = NewMariadbConnection(*cp)
	case "postgresql":
		conn = NewPostgresqlConnection(*cp)
	default:
		break
	}
	db, err = conn.ConnectDb()
	return db, err
}

func (cp *ConnectionProfile) SyncTagetDbSchema(targetDB *gorm.DB) error {
	var err error
	var conn DbConnecter

	switch strings.ToLower(cp.Dbms) {
	case "mysql":
		conn = NewMysqlConnection(*cp)
	case "mariadb":
		conn = NewMariadbConnection(*cp)
	case "postgresql":
		conn = NewPostgresqlConnection(*cp)
	default:
		break
	}
	err = conn.SyncSchema(targetDB)
	return err
}

func (cp *ConnectionProfile) conenctSrcDb() (*gorm.DB, error) {
	var srcDb *gorm.DB
	srcDbDsn := os.Getenv("MARIADB_DSN")
	srcDb, err := gorm.Open(mysql.Open(srcDbDsn), &gorm.Config{})
	if err != nil {
		log.Println("######## Failed to connect to source database ########\n", err.Error())
		return nil, err
	}
	return srcDb, nil
}
