package middleware

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var hmacSecretKey []byte

func JWTAuthen() gin.HandlerFunc {
	return func(c *gin.Context) {
		//get token from header
		hmacSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
		header := c.GetHeader("Authorization")
		tokenString := header[7:]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("there was an error")
			}
			return hmacSecretKey, nil
		})

		//check token claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Set variable into Context
			c.Set("userId", claims["ID"])
		} else {
			c.AbortWithStatusJSON(401, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
		}
		c.Next()
	}
}

func JWTAuthenAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {

		//get token from header
		hmacSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
		header := c.GetHeader("Authorization")
		tokenString := header[7:]
		
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("there was an error")
			}
			return hmacSecretKey, nil
		})
		
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("userId", claims["ID"])
			if claims["Role"] != "admin" {
				log.Println("Not Data Admin")
				c.AbortWithStatusJSON(401, gin.H{
					"status":  "error",
					"message": "You are not data admin",
				})
			}
		} else {
			log.Println("####### JWT claim failed #######\n", err.Error())
			c.AbortWithStatusJSON(401, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
		}
		c.Next()
	}
}
