package connectionprofilecontroller

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"masterdb/database"
	"masterdb/models"
)

// ********************* GET ALL *********************
func GetAll(c *gin.Context) {
	// Get all records
	var connectionProfile []models.ConnectionProfile
	database.Db.Find(&connectionProfile)

	//hide sensitive data
	for i := 0; i < len(connectionProfile); i++ {
		models.HideSensitiveData(&connectionProfile[i])
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Gell All Connection Profiles",
		"data":    connectionProfile,
	})
}

// ********************* GET BY USER ID *********************
func GetAllByUserId(c *gin.Context) {
	userId := c.MustGet("userId")

	var connectionProfile []models.ConnectionProfile
	database.Db.Where("user_id = ?", userId).Preload("ImportedDatabase").Find(&connectionProfile)

	for i := 0; i < len(connectionProfile); i++ {
		models.HideSensitiveData(&connectionProfile[i])
		decryptedPassword := connectionProfile[i].GetDecryptPassword()
		connectionProfile[i].Password = decryptedPassword
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Get All Connection Profiles by user id",
		"data":    connectionProfile,
	})
}

// ********************* GET ONE *********************
func GetOne(c *gin.Context) {
	//get connection profile id from url
	connectionProfileId := c.Query("id")

	var connectionProfile models.ConnectionProfile
	result := database.Db.First(&connectionProfile, connectionProfileId)
	database.Db.Preload("ImportedDatabase").Find(&connectionProfile)

	if result.Error == nil && connectionProfile.ID > 0 {
		models.HideSensitiveData(&connectionProfile)
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Connection Profile found",
			"data":    connectionProfile,
		})
		return
	} else {
		models.HideSensitiveData(&connectionProfile)
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Connection Profile not found" + result.Error.Error(),
			"data":    connectionProfile,
		})
		return
	}
}

// ********************* CREATE *********************
type CreateBody struct {
	Dbms           string `json:"dbms" binding:"required"`
	ConnectionName string `json:"connectionName" binding:"required"`
	Host           string `json:"host" binding:"required"`
	Port           string `json:"port" binding:"required"`
	DatabaseName   string `json:"databaseName" binding:"required"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}

func Create(c *gin.Context) {
	//Create new connection profile
	userId := c.MustGet("userId")

	var json CreateBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error invalid request #######\n", err.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error invalid request",
		})
		return
	}
	var connectionProfile models.ConnectionProfile

	//check if connection_name already exists
	database.Db.Where("connection_name = ? AND user_id = ?", json.ConnectionName, userId).First(&connectionProfile)
	if connectionProfile.ID > 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Connection name already exists",
		})
		return
	}

	//check if user exists
	var user models.User
	database.Db.First(&user, userId)
	if user.ID == 0 {
		log.Println("******* User not found *******")
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "User not found",
		})
		return
	}

	//create connection profile
	connectionProfile = models.ConnectionProfile{
		Dbms:           json.Dbms,
		ConnectionName: json.ConnectionName,
		Host:           json.Host,
		Port:           json.Port,
		DatabaseName:   json.DatabaseName,
		Username:       json.Username,
		Password:       json.Password,
		UserID:         uint(int(userId.(float64))),
		ImportedDatabase: models.ImportedDatabase{
			Name:   json.DatabaseName,
			Dbms:   json.Dbms,
			Status: "pending",
			UserID: uint(int(userId.(float64))),
		},
	}
	//encrypt password
	connectionProfile.EncryptPassword()
	result := database.Db.Create(&connectionProfile)
	if result.RowsAffected > 0 && connectionProfile.ID > 0 && result.Error == nil {
		//associate connection profile with user
		var user models.User
		database.Db.First(&user, userId)
		database.Db.Model(&user).Association("ConnectionProfiles").Append(&connectionProfile)
		database.Db.Save(&user)

		database.Db.Model(&user).Association("ImportedDatabases").Append(&connectionProfile.ImportedDatabase)
		database.Db.Save(&user)

		models.HideSensitiveData(&connectionProfile)

		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Connection Profile created",
			"data":    connectionProfile,
		})
		return
	} else {
		models.HideSensitiveData(&connectionProfile)
		log.Println("####### Connection Profile not created #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Connection Profile not created",
			"data":    connectionProfile,
		})
		return
	}
}

// ********************* UPDATE *********************
type UpdateBody struct {
	Dbms           string `json:"dbms" binding:"required"`
	ConnectionName string `json:"connectionName" binding:"required"`
	Host           string `json:"host" binding:"required"`
	Port           string `json:"port" binding:"required"`
	DatabaseName   string `json:"databaseName" binding:"required"`
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password"`
}

func UpdateOne(c *gin.Context) {
	//Update connection profile
	//get connection profile id from url
	connectionProfileId := c.Query("id")

	var json UpdateBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error in json body #######\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error in json body",
		})
		return
	}
	var connectionProfile models.ConnectionProfile
	
	// check if connection_name already exists
	database.Db.Where("connection_name = ?", json.ConnectionName).First(&connectionProfile)
	uintId, _ := strconv.ParseUint(connectionProfileId, 10, 32)
	if connectionProfile.ID > 0 && connectionProfile.ID != uint(uintId) {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Connection name already exists",
		})
		return
	}

	//check if connection profile exists
	result := database.Db.First(&connectionProfile, connectionProfileId)
	if result.Error != nil && result.RowsAffected <= 0 && connectionProfile.ID <= 0 {
		models.HideSensitiveData(&connectionProfile)

		log.Println("####### Connection Profile not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Connection Profile not found",
			"data":    connectionProfile,
		})
		return
	}

	//update connection profile
	if json.Dbms != "" {
		connectionProfile.Dbms = json.Dbms
	}
	if json.ConnectionName != "" {
		connectionProfile.ConnectionName = json.ConnectionName
	}
	if json.Host != "" {
		connectionProfile.Host = json.Host
	}
	if json.Port != "" {
		connectionProfile.Port = json.Port
	}
	if json.DatabaseName != "" {
		connectionProfile.DatabaseName = json.DatabaseName
	}
	if json.Username != "" {
		connectionProfile.Username = json.Username
	}
	if json.Password != "" {
		connectionProfile.Password = json.Password
	}
	//encrypt password
	connectionProfile.EncryptPassword()
	database.Db.Save(&connectionProfile)
	database.Db.Preload("ImportedDatabase").First(&connectionProfile)

	models.HideSensitiveData(&connectionProfile)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Connection Profile update successful",
		"data":    connectionProfile,
	})
}

// ********************* DELETE *********************
func DeleteOne(c *gin.Context) {
	//Delete connection profile
	//get connection profile id from url
	connectionProfileId := c.Query("id")

	var connectionProfile models.ConnectionProfile
	result := database.Db.First(&connectionProfile, connectionProfileId)

	if result.Error == nil && result.RowsAffected > 0 && connectionProfile.ID > 0 {
		//delete connection profile
		database.Db.Unscoped().Delete(&connectionProfile)
		// database.Db.Delete(&connectionProfile)
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Connection Profile deleted",
		})
	} else {
		log.Println("####### Connection Profile not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Connection Profile not found",
		})
	}
}

// ********************* IMPORT DATABASE *********************
type ImportDbBody struct {
	DatabaseName string `json:"databaseName" binding:"required"`
}

func ImportDatabaseByConnectionId(c *gin.Context) {
	//get connection profile id from url
	connectionProfileId := c.Query("id")

	var json ImportDbBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var connectionProfile models.ConnectionProfile
	result := database.Db.Preload("ImportedDatabase").First(&connectionProfile, connectionProfileId)

	if result.Error != nil && result.RowsAffected <= 0 && connectionProfile.ID <= 0 {
		log.Println("####### Connection Profile not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Connection Profile not found",
		})
		return
	}
	//check if connectionProfile has already been imported
	if connectionProfile.ImportedDatabase.ID > 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Database already imported",
		})
		return
	}

	//create imported database
	importedDatabase := models.ImportedDatabase{
		Name:                json.DatabaseName,
		Dbms:                connectionProfile.Dbms,
		Status:              "pending",
		UserID:              connectionProfile.UserID,
		ConnectionProfileID: connectionProfile.ID,
	}

	if result := database.Db.Create(&importedDatabase); result.Error != nil {
		log.Println("####### Imported Database not created #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Imported Database not created",
		})
		return
	}

	//update connection profile
	connectionProfile.ImportedDatabase = importedDatabase

	models.HideSensitiveData(&connectionProfile)
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Imported Database Success",
		"data":    connectionProfile,
	})
}
