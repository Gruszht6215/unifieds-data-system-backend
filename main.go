package main

import (
	"log"
	// "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	AuthController "masterdb/controllers/authcontroller"
	ClusterController "masterdb/controllers/clustercontroller"
	ColumnController "masterdb/controllers/columncontroller"
	ConnectionProfileController "masterdb/controllers/connectionprofilecontroller"
	ImportedDatabaseController "masterdb/controllers/importeddatabasecontroller"
	TableController "masterdb/controllers/tablecontroller"
	TagController "masterdb/controllers/tagcontroller"
	UserController "masterdb/controllers/usercontroller"
	DashboardController "masterdb/controllers/dashboardcontroller"
	"masterdb/database"
	"masterdb/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	database.InitDB()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	// r.Use(cors.Default())

	r.POST("/register", AuthController.RegisterAdmin)
	r.POST("/login", AuthController.Login)

	authorized := r.Group("/api", middleware.JWTAuthen())
	authorized.GET("/user/getAll", UserController.GetAll)
	authorized.GET("/user/getOne", UserController.GetOne)
	authorized.GET("/dashboard/getTagColumnTreeMapData", DashboardController.GetTagColumnTreeMapData)

	// ############################## DATA ADMIN ##############################
	authorizedDataAdmin := r.Group("/api", middleware.JWTAuthenAdmin())
	authorizedDataAdmin.DELETE("/user/deleteOne", UserController.DeleteOne)

	authorizedDataAdmin.GET("/connection-profile/getAll", ConnectionProfileController.GetAll)
	authorizedDataAdmin.GET("/connection-profile/getAllByUserId", ConnectionProfileController.GetAllByUserId)
	authorizedDataAdmin.GET("/connection-profile/getOne", ConnectionProfileController.GetOne)
	authorizedDataAdmin.POST("/connection-profile/create", ConnectionProfileController.Create)
	authorizedDataAdmin.PUT("/connection-profile/updateOne", ConnectionProfileController.UpdateOne)
	authorizedDataAdmin.DELETE("/connection-profile/deleteOne", ConnectionProfileController.DeleteOne)
	authorizedDataAdmin.POST("/connection-profile/importDatabaseByConnectionId", ConnectionProfileController.ImportDatabaseByConnectionId)

	authorizedDataAdmin.POST("/imported-database/sync-schema", ImportedDatabaseController.SyncSchema)
	authorizedDataAdmin.GET("/imported-database/getAllByUserId", ImportedDatabaseController.GetAllByUserId)
	authorizedDataAdmin.DELETE("/imported-database/deleteOne", ImportedDatabaseController.Delete)
	authorizedDataAdmin.PUT("/imported-database/updateOne", ImportedDatabaseController.UpdateOne)

	authorizedDataAdmin.GET("/table/getAllByImportedDbId", TableController.GetAllByImportedDbId)
	authorizedDataAdmin.PUT("/table/updateOne", TableController.UpdateOne)

	authorizedDataAdmin.GET("/column/getAllByTableId", ColumnController.GetAllByTableId)
	authorizedDataAdmin.PUT("/column/updateOne", ColumnController.UpdateOne)
	authorizedDataAdmin.PUT("/column/updateTagsColumnByColumnId", ColumnController.UpdateTagsByColumnId)

	authorizedDataAdmin.GET("/cluster/getAllByUserId", ClusterController.GetAllByUserId)
	authorizedDataAdmin.POST("/cluster/create", ClusterController.Create)
	authorizedDataAdmin.PUT("/cluster/updateTagsClusterByClusterId", ClusterController.UpdateTagsByClusterId)
	authorizedDataAdmin.PUT("/cluster/updateOne", ClusterController.UpdateOne)
	authorizedDataAdmin.DELETE("/cluster/deleteOne", ClusterController.DeleteOne)

	authorizedDataAdmin.GET("/tag/getAllByUserId", TagController.GetAllByUserId)
	authorizedDataAdmin.POST("/tag/create", TagController.Create)
	authorizedDataAdmin.PUT("/tag/updateOne", TagController.UpdateOne)


	r.Run("localhost:8080")
}
