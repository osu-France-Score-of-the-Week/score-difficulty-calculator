package main

import (
	"log"
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
			log.Printf("[calc] bind failed err=%v", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		log.Printf(
			"[calc] request beatmap=%d version=%q cs=%.2f od=%.2f ar=%.2f star=%.3f max_combo=%d score=%d acc=%.6f pp=%.2f mods=%v stats=%+v",
			req.Beatmap.ID,
			req.Beatmap.Version,
			req.Beatmap.CS,
			req.Beatmap.OD,
			req.Beatmap.AR,
			req.BeatmapAttributes.StarRating,
			req.BeatmapAttributes.MaxCombo,
			req.Score.ID,
			req.Score.Accuracy,
			req.Score.PP,
			req.Score.Mods,
			req.Score.Statistics,
		)

		result, detail := calculator.Compute(req.BeatmapAttributes, req.Score, req.Beatmap)
		log.Printf("[calc] result beatmap=%d score=%d final=%.6f aim=%.6f speed=%.6f map=%.6f acc=%.6f miss_pen=%.6f slider=%.6f ar_mult=%.6f eff_ar=%.6f eff_od=%.6f misses=%d",
			req.Beatmap.ID,
			req.Score.ID,
			result,
			detail.AimScore,
			detail.SpeedScore,
			detail.MapScore,
			detail.AccMult,
			detail.MissPenalty,
			detail.SliderScore,
			detail.ARMult,
			detail.EffectiveAR,
			detail.EffectiveOD,
			detail.MissCount,
		)
		c.JSON(200, result)
	})

	r.Run(":8180")
}
