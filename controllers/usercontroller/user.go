package usercontroller

import (
	"github.com/gin-gonic/gin"

	"masterdb/database"
	"masterdb/models"
)

var hmacSecretKey []byte

func GetAll(c *gin.Context) {
	// Get all records
	var users []models.User
	database.Db.Find(&users)
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Gell All Users",
		"data":    users,
	})
}

func GetOne(c *gin.Context) {
	// Get one record
	userId := c.MustGet("userId")

	var user models.User
	database.Db.Preload("ConnectionProfiles").Preload("ImportedDatabases").First(&user, userId)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Get One User",
		"data":    user,
	})
}

func DeleteOne(c *gin.Context) {
	// On test will change later
	// Delete one record
	userId := c.MustGet("userId")

	var user models.User
	database.Db.First(&user, userId)
	database.Db.Unscoped().Delete(&user, userId)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Delete One User",
		"data":    user,
	})
}
