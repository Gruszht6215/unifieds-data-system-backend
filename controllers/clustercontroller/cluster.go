package clustercontroller

import (
	"log"

	"github.com/gin-gonic/gin"

	// "net/http"

	"masterdb/database"
	"masterdb/models"
)

// ********************* GET ALL BY USER ID *********************
func GetAllByUserId(c *gin.Context) {
	//get connection profile id from url
	userId := c.MustGet("userId")

	var clusters []models.Cluster
	result := database.Db.Where("user_id = ?", userId).Preload("Tags").Find(&clusters)
	if result.Error != nil {
		log.Println("####### Cluster not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Cluster not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Cluster found",
		"data":    clusters,
	})
}

// ********************* CREATE *********************
type CreateBody struct {
	Name string `json:"name" binding:"required"`
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

	var cluster models.Cluster
	result := database.Db.Where("name = ? AND user_id = ?", json.Name, userId).Find(&cluster)
	if result.RowsAffected > 0 && cluster.ID > 0 {
		log.Println("####### Cluster name already exists #######")
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Cluster name already exists",
		})
		return
	}

	cluster.Name = json.Name
	cluster.UserID = uint(userId.(float64))
	result = database.Db.Create(&cluster)
	if result.Error != nil {
		log.Println("####### Error creating cluster #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error creating cluster",
		})
		return
	}

	//preload tags
	database.Db.Preload("Tags").Find(&cluster)
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Cluster created",
		"data":    cluster,
	})
}

// ********************* UPDATE TAGS ASSOCIATION BY CLUSTER ID*********************
type UpdateAppendTagsBody struct {
	TagIDs []uint `json:"tagIds" binding:"required"`
}

func UpdateTagsByClusterId(c *gin.Context) {
	userId := c.MustGet("userId")
	clusterId := c.Query("id")

	var json UpdateAppendTagsBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error invalid request #######\n", err.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error invalid request",
		})
		return
	}

	var cluster models.Cluster
	result := database.Db.Where("id = ? AND user_id = ?", clusterId, userId).Find(&cluster)
	if result.Error != nil {
		log.Println("####### Cluster not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Cluster not found",
		})
		return
	}

	
	
	var tags []models.Tag
	//check if tagids is empty
	if len(json.TagIDs) == 0 {
		database.Db.Model(&cluster).Association("Tags").Clear()
		database.Db.Preload("Tags").Find(&cluster)
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Cluster updated",
			"data":    cluster,
		})
		return
	}
	result = database.Db.Where("id IN (?) AND user_id = ?", json.TagIDs, userId).Find(&tags)
	if result.Error != nil {
		log.Println("####### Tags not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Tags not found",
		})
		return
	}

	database.Db.Model(&cluster).Association("Tags").Replace(&tags)
	
	result = database.Db.Save(&cluster)
	if result.Error != nil {
		log.Println("####### Error updating cluster #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error updating cluster",
		})
		return
	}

	database.Db.Preload("Tags").Find(&cluster)
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Cluster updated",
		"data":    cluster,
	})
}

// ********************* UPDATE CLUSTER *********************
type UpdateBody struct {
	Name string `json:"name" binding:"required"`
}

func UpdateOne(c *gin.Context) {
	userId := c.MustGet("userId")
	clusterId := c.Query("id")

	var json UpdateBody
	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println("####### Error invalid request #######\n", err.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error invalid request",
		})
		return
	}

	var cluster models.Cluster
	result := database.Db.Where("id = ? AND user_id = ?", clusterId, userId).Find(&cluster)
	if result.Error != nil {
		log.Println("####### Cluster not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Cluster not found",
		})
		return
	}

	cluster.Name = json.Name
	result = database.Db.Save(&cluster)
	database.Db.Preload("Tags").Find(&cluster)
	if result.Error != nil {
		log.Println("####### Error updating cluster #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error updating cluster",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Cluster updated",
		"data":    cluster,
	})
}

// ********************* DELETE CLUSTER *********************
func DeleteOne(c *gin.Context) {
	userId := c.MustGet("userId")
	clusterId := c.Query("id")

	var cluster models.Cluster
	result := database.Db.Where("id = ? AND user_id = ?", clusterId, userId).Find(&cluster)
	if result.Error != nil {
		log.Println("####### Cluster not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Cluster not found",
		})
		return
	}

	result = database.Db.Select("Tags").Unscoped().Delete(&cluster)
	if result.Error != nil {
		log.Println("####### Error deleting cluster #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Error deleting cluster",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Cluster deleted",
	})
}