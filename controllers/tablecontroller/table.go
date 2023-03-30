package tablecontroller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"masterdb/database"
	"masterdb/models"
)

// ********************* GET ALL BY IMPORTED DB ID *********************
func GetAllByImportedDbId(c *gin.Context) {
	//get connection profile id from url
	improtedDbId := c.Query("id")

	var tables []models.Table
	result := database.Db.Where("imported_database_id = ?", improtedDbId).Preload("Columns").Preload("Columns.Tags").Find(&tables)
	if result.Error != nil && result.RowsAffected <= 0 && len(tables) <= 0 {
		log.Println("####### Column not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Column not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Column found",
		"data":    tables,
	})
}

// ********************* UPDATE BY TABLE ID*********************
type UpdateBody struct {
	Description string `json:"description" binding:"required"`
}

func UpdateOne(c *gin.Context) {
	tableId := c.Query("id")

	var json UpdateBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error invalid request #######\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"meesage": "Error invalid request",
		})
		return
	}

	var table models.Table
	result := database.Db.Where("id = ?", tableId).Find(&table)
	if result.Error != nil && result.RowsAffected <= 0 && table.ID <= 0 {
		log.Println("####### Table not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Table not found",
		})
		return
	}

	result = database.Db.Model(&table).Update("description", json.Description)

	if result.Error != nil {
		log.Println("####### Error updating table #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error updating table",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Table updated",
		"data":    table,
	})
}
