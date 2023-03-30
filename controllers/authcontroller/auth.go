package authcontroller

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	"masterdb/database"
	"masterdb/models"
)

// var hmacSecretKey []byte = []byte("secret")
var hmacSecretKey []byte

// ********************* REGISTER *********************
// bindings from JSON body
type RegisterBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
	Host     string `json:"host" binding:"required"`
}

func RegisterAdmin(c *gin.Context) {
	var json RegisterBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var user models.User

	//check if username already exists
	database.Db.Where("username = ?", json.Username).First(&user)
	if user.ID > 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "username already exists",
		})
		return
	}

	//hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(json.Password), bcrypt.DefaultCost)

	//create user
	user = models.User{
		Username: json.Username,
		Password: string(hashedPassword),
		Role:     json.Role,
		Host:     json.Host,
	}
	database.Db.Create(&user)
	if user.ID > 0 {
		query := "CREATE USER '" + user.Username + "'@'" + user.Host + "' IDENTIFIED BY '" + json.Password + "'"
		if err := database.Db.Exec(query).Error; err != nil {
			log.Println("######## CREATE USER ERROR ######\n", err)
			database.Db.Unscoped().Delete(&user)
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "user not created",
			})
			return
		}

		query = "GRANT ALL PRIVILEGES ON *.* TO '" + user.Username + "'@'" + user.Host + "'IDENTIFIED BY '" + json.Password + "' WITH GRANT OPTION"
		if err := database.Db.Exec(query).Error; err != nil {
			log.Println("######## GRANT USER ERROR #######\n", err)
			database.Db.Unscoped().Delete(&user)
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "user can not grant privileges",
			})
			return
		}
		database.Db.Exec("FLUSH PRIVILEGES")
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "user created",
			"userID":  user.ID,
		})
	} else {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "user not created",
		})
	}
}

// ********************* LOGIN *********************
type LoginBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var json LoginBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//check if user exists
	var user models.User
	database.Db.Where("username = ?", json.Username).First(&user)
	// database.Db.Preload("ConnectionProfiles").Find(&user)

	if user.ID == 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "user not found",
		})
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(json.Password))
	if err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "invalid password",
		})
		return
	} else {
		hmacSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"userId": user.ID,
			"role":   user.Role,
			"exp":    time.Now().Add(time.Hour * 2).Unix(),
		})
		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := token.SignedString(hmacSecretKey)
		log.Println(tokenString, err)

		user.Password = ""
		// user.ID = 0
		type userBody struct {
			ID       uint
			Username string
			Password string
			Role     string // admin, consumer
			Host     string
			Token    string
		}
		var userRes = userBody{
			ID:       user.ID,
			Username: user.Username,
			Password: user.Password,
			Role:     user.Role,
			Host:     user.Host,
			Token:    tokenString,
		}
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "logged in",
			"token":   tokenString,
			"user":    userRes,
		})
	}
}
