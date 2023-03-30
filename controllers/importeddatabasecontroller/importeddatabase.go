package importeddatabasecontroller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"masterdb/database"
	"masterdb/models"
)

// ********************* SYNC SCHEMA *********************
func SyncSchema(c *gin.Context) {
	//Sync database schema
	//get imported database id from url
	importedDatabaseId := c.Query("id")

	var importedDatabase models.ImportedDatabase
	result := database.Db.First(&importedDatabase, importedDatabaseId)
	if result.Error != nil && result.RowsAffected <= 0 && importedDatabase.ID <= 0 {
		log.Println("####### Imported Database not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Imported Database not found",
		})
		return
	}

	//get connection profile
	var connectionProfile models.ConnectionProfile
	if database.Db.First(&connectionProfile, importedDatabase.ConnectionProfileID).Error != nil {
		log.Println("####### Connection Profile not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Connection Profile not found. connection profile may be deleted.",
		})
		return
	}

	var targetDB *gorm.DB
	var err error
	// Get currrent user to check if this user have privileges to access
	userId := c.MustGet("userId")
	var user models.User
	database.Db.First(&user, userId)
	if user.ID == 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "User not found",
		})
		return
	}
	
	targetDB, err = connectionProfile.ConnectTargetDb()
	if err != nil {
		log.Println("######## Failed to connect to targer database ########\n", err)
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Failed to connect to targer database ",
		})
		return
	}
	//delete all columns that associated with this database's tables
	database.Db.Where("table_id IN (SELECT id FROM tables WHERE imported_database_id = ?)", importedDatabase.ID).Unscoped().Delete(&models.Column{})
	//delete all tables that associated with this imported database
	database.Db.Where("imported_database_id = ?", importedDatabase.ID).Unscoped().Delete(&models.Table{})

	if connectionProfile.SyncTagetDbSchema(targetDB) != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Failed to sync schema",
		})
		return
	}

	//update imported database status
	database.Db.Model(&importedDatabase).Update("status", "active")

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Sync Schema Success",
	})
}

// ********************* GET BY USER ID *********************
func GetAllByUserId(c *gin.Context) {
	userId := c.MustGet("userId")
	var importedDatabases []models.ImportedDatabase
	database.Db.Where("user_id = ?", userId).Preload("Tables").Preload("Tables.Columns").Preload("Tables.Columns.Tags").Find(&importedDatabases)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Get All Imported database by user id",
		"data":    importedDatabases,
	})
	return
}

// ********************* DELETE *********************
func Delete(c *gin.Context) {
	importedDbId := c.Query("id")

	var importedDatabase models.ImportedDatabase
	database.Db.First(&importedDatabase, importedDbId)
	if importedDatabase.ID == 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Imported database not found",
		})
		return
	}
	recordResult := database.Db.Unscoped().Delete(&importedDatabase)

	if recordResult.Error != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Delete imported database failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Delete imported database success",
	})
}

// ********************* UPDATE BY DB TABLE ID*********************
type UpdateBody struct {
	Description string `json:"description"`
}

func UpdateOne(c *gin.Context) {
	importedDbId := c.Query("id")

	var json UpdateBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error invalid request #######\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error invalid request",
		})
		return
	}

	var importedDatabase models.ImportedDatabase
	database.Db.First(&importedDatabase, importedDbId)
	if importedDatabase.ID == 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Database not found",
		})
		return
	}

	importedDatabase.Description = json.Description
	database.Db.Save(&importedDatabase)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Update imported database success",
		"data":    importedDatabase,
	})
}
