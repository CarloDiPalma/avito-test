package utils

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetDB(c *gin.Context) (*gorm.DB, bool) {
	db, exists := c.Get("db")
	if !exists {
		c.JSON(500, gin.H{"reason": "database connection not found"})
		return nil, false
	}

	database, ok := db.(*gorm.DB)
	if !ok {
		c.JSON(500, gin.H{"reason": "invalid database connection"})
		return nil, false
	}

	return database, true
}
