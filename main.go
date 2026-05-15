package main

import (
	"score-difficulty-calculator/calculator"
	"score-difficulty-calculator/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	r.POST("/analyze", func(c *gin.Context) {
		var req models.AnalyzeRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		result, _ := calculator.Compute(req.BeatmapAttributes, req.Score, req.Beatmap)
		c.JSON(200, result)
	})

	r.Run(":8180")
}
