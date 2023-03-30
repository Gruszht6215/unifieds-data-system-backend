package dashboardcontroller

import (
	"github.com/gin-gonic/gin"
	"log"
	"masterdb/database"
	"masterdb/models"
)

// ********************* GET tag-column treemap data *********************
func GetTagColumnTreeMapData(c *gin.Context) {
	userId := c.MustGet("userId")

	//get all tags by user id
	var tags []models.Tag
	result := database.Db.Where("user_id = ?", userId).Preload("Columns").Find(&tags)
	if result.Error != nil {
		log.Println("####### Tag not found #######\n", result.Error.Error())
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Tag not found",
		})
		return
	}

	var tagColumnCount []map[string]interface{}
	for i := 0; i < len(tags); i++ {
		tagColumnCount = append(tagColumnCount, map[string]interface{}{
			"name":  tags[i].Name,
			"columnAmount": len(tags[i].Columns),
		})
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Tag found",
		"data":    tagColumnCount,
	})
}
