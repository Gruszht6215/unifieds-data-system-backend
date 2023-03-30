package tagcontroller

import (
	"log"

	"github.com/gin-gonic/gin"

	// "github.com/icza/gog"

	"masterdb/database"
	"masterdb/models"
)

// ********************* GET ALL BY USER ID *********************
func GetAllByUserId(c *gin.Context) {
	//get connection profile id from url
	userId := c.MustGet("userId")

	var tags []models.Tag
	result := database.Db.Where("user_id = ?", userId).Preload("Columns").Preload("Clusters").Find(&tags)
	if result.Error != nil {
		log.Println("####### Tag not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Tag not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Tag found",
		"data":    tags,
	})
}

// ********************* CREATE *********************
type CreateBody struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	ClusterID   uint   `json:"clusterId"`
	ColumnID    uint   `json:"columnId"`
}

func Create(c *gin.Context) {
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

	var tag models.Tag
	result := database.Db.Where("name = ? AND user_id = ?", json.Name, userId).Find(&tag)
	if result.RowsAffected > 0 && tag.ID > 0 {
		log.Println("####### Tag name already exists #######")
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Tag name already exists",
		})
		return
	}

	tag = models.Tag{
		Name:        json.Name,
		Description: json.Description,
		Color:       json.Color,
		UserID:      uint(userId.(float64)),
	}

	result = database.Db.Create(&tag)
	if result.Error != nil {
		log.Println("####### Error creating tag #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error creating tag",
		})
		return
	}

	//check if user pass column id
	if json.ColumnID > 0 {
		var column models.Column
		result = database.Db.Where("id = ?", json.ColumnID).Find(&column)
		if result.Error != nil {
			log.Println("####### Error finding column #######\n", result.Error.Error())
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Column not found",
			})
			return
		}
		database.Db.Model(&column).Association("Tags").Append(&tag)
	}

	//check if user pass cluster id
	if json.ClusterID > 0 {
		var cluster models.Cluster
		result = database.Db.Where("id = ?", json.ClusterID).Find(&cluster)
		if result.Error != nil {
			log.Println("####### Error finding cluster #######\n", result.Error.Error())
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Cluster not found",
			})
			return
		}
		database.Db.Model(&cluster).Association("Tags").Append(&tag)
	}

	//preload
	database.Db.Preload("Columns").Preload("Clusters").First(&tag)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Tag created",
		"data":    tag,
	})
}

// ********************* UPDATE BY TAG ID*********************
type UpdateBody struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func UpdateOne(c *gin.Context) {
	userId := c.MustGet("userId")
	tagId := c.Query("id")

	var json UpdateBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error invalid request #######\n", err.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error invalid request",
		})
		return
	}

	var tag models.Tag
	result := database.Db.Where("id = ? AND user_id = ?", tagId, userId).Find(&tag)
	if result.Error != nil {
		log.Println("####### Tag not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Tag not found",
		})
		return
	}

	tag.Name = json.Name
	tag.Description = json.Description
	tag.Color = json.Color

	result = database.Db.Save(&tag)
	if result.Error != nil {
		log.Println("####### Error updating tag #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error updating tag",
		})
		return
	}

	database.Db.Preload("Columns").Preload("Clusters").First(&tag)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Tag updated",
		"data":    tag,
	})
}