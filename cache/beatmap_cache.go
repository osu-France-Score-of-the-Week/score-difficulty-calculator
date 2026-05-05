package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"score-difficulty-calculator/models"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// CacheEntry represents a single cached beatmap + mod combination.
type CacheEntry struct {
	BeatmapID int
	Mods      []string
}

// BeatmapCache is a thread-safe, file-backed cache for BeatmapAttributes.
// Entries are keyed by beatmap ID and a sorted, comma-joined list of mods,
// e.g. "12345_DT,HD" or "12345_" when no mods are applied.
type BeatmapCache struct {
	mu   sync.RWMutex
	data map[string]models.BeatmapAttributes
	path string
}

// NewBeatmapCache creates a BeatmapCache backed by the JSON file at path.
// If the file exists its contents are loaded into memory; if it does not
// exist the cache starts empty and the file will be created on the first
// successful Set call.
func NewBeatmapCache(path string) (*BeatmapCache, error) {
	c := &BeatmapCache{
		data: make(map[string]models.BeatmapAttributes),
		path: path,
	}

	if _, err := os.Stat(path); err == nil {
		file, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("beatmap cache: read file %q: %w", path, err)
		}
		if err := json.Unmarshal(file, &c.data); err != nil {
			return nil, fmt.Errorf("beatmap cache: parse file %q: %w", path, err)
		}
	}

	return c, nil
}

// cacheKey builds a stable lookup key from a beatmap ID and an arbitrary
// list of mods.  Mods are sorted before joining so that ["HD","DT"] and
// ["DT","HD"] produce the same key.
func cacheKey(beatmapID int, mods []string) string {
	sorted := make([]string, len(mods))
	copy(sorted, mods)
	sort.Strings(sorted)
	return fmt.Sprintf("%d_%s", beatmapID, strings.Join(sorted, ","))
}

// Get returns the cached BeatmapAttributes for the given beatmap ID and
// mods.  The second return value is false when no entry exists.
func (c *BeatmapCache) Get(beatmapID int, mods []string) (models.BeatmapAttributes, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	attrs, ok := c.data[cacheKey(beatmapID, mods)]
	return attrs, ok
}

// GetAll returns all cached entries as a slice of CacheEntry,
// each containing the beatmap ID and its associated mods.
func (c *BeatmapCache) GetAll() []CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]CacheEntry, 0, len(c.data))
	for key := range c.data {
		// key format: "<beatmapID>_<MOD1,MOD2,...>" or "<beatmapID>_"
		parts := strings.SplitN(key, "_", 2)
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		var mods []string
		if len(parts) == 2 && parts[1] != "" {
			mods = strings.Split(parts[1], ",")
		}
		entries = append(entries, CacheEntry{BeatmapID: id, Mods: mods})
	}
	return entries
}

// Set stores attributes in the in-memory cache and immediately persists the
// entire cache to disk.  An error is returned only when the file cannot be
// written; the in-memory entry is always stored regardless.
func (c *BeatmapCache) Set(beatmapID int, mods []string, attributes models.BeatmapAttributes) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[cacheKey(beatmapID, mods)] = attributes

	encoded, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return fmt.Errorf("beatmap cache: marshal: %w", err)
	}

	if err := os.WriteFile(c.path, encoded, 0644); err != nil {
		return fmt.Errorf("beatmap cache: write file %q: %w", c.path, err)
	}

	return nil
}
