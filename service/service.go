package service

import "score-difficulty-calculator/cache"

type OsuService struct {
	Cache *cache.BeatmapCache
}

func NewOsuService(c *cache.BeatmapCache) *OsuService {
	return &OsuService{Cache: c}
}
