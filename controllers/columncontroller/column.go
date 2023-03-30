package columncontroller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"masterdb/database"
	"masterdb/models"
)

// ********************* GET ALL BY TABLE ID *********************
func GetAllByTableId(c *gin.Context) {
	//get connection profile id from url
	tableId := c.Query("id")

	var dbTables []models.Column
	result := database.Db.Where("table_id = ?", tableId).Find(&dbTables)
	if result.Error != nil {
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
		"data":    dbTables,
	})
}

// ********************* UPDATE BY COLUMN ID*********************
type UpdateBody struct {
	Description string `json:"description"`
}

func UpdateOne(c *gin.Context) {
	dbTableId := c.Query("id")

	var json UpdateBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error invalid request #######\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error invalid request",
		})
		return
	}

	var column models.Column
	result := database.Db.Where("id = ?", dbTableId).Find(&column)
	if result.Error != nil || result.RowsAffected <= 0 || column.ID <= 0 {
		log.Println("####### Column not found #######")
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Column not found",
		})
		return
	}

	database.Db.Preload("Tags").Find(&column)

	column.Description = json.Description
	database.Db.Save(&column)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Update Column success",
		"data":    column,
	})
}

// ********************* UPDATE TAGS ASSOCIATION BY COLUMN ID*********************
type UpdateTagBody struct {
	TagIDs []uint `json:"tagIds" binding:"required"`
}

func UpdateTagsByColumnId(c *gin.Context) {
	columnId := c.Query("id")

	var json UpdateTagBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error invalid request #######\n", err.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error invalid request",
		})
		return
	}

	var column models.Column
	result := database.Db.Where("id = ?", columnId).Find(&column)
	if result.Error != nil {
		log.Println("####### Column not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Column not found",
		})
		return
	}

	var tags []models.Tag
	if len(json.TagIDs) == 0 {
		database.Db.Model(&column).Association("Tags").Clear()
		database.Db.Preload("Tags").Find(&column)
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Column updated",
			"data":    column,
		})
		return
	} else {
		result = database.Db.Where("id IN (?)", json.TagIDs).Find(&tags)
		if result.Error != nil || result.RowsAffected == 0 || len(tags) == 0 {
			log.Println("####### Tag not found #######")
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Tag not found",
			})
			return
		}
		database.Db.Model(&column).Association("Tags").Replace(&tags)
		database.Db.Preload("Tags").Find(&column)
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Column updated",
		"data":    column,
	})
}
