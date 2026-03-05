package kumi

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ── Data model ───────────────────────────────────

// LauncherID identifies a launcher for cache keying.
type LauncherID string

const (
	LauncherMinecraft  LauncherID = "minecraft_launcher"
	LauncherMultiMC    LauncherID = "multimc"
	LauncherCurseForge LauncherID = "curseforge"
	LauncherModrinth   LauncherID = "modrinth"
	LauncherPrism      LauncherID = "prismlauncher"
	LauncherATLauncher LauncherID = "atlauncher"
	LauncherGDLauncher LauncherID = "gdlauncher"
	LauncherTechnic    LauncherID = "technic"
	LauncherPolyMC     LauncherID = "polymc"
	LauncherFeather    LauncherID = "feather"
	LauncherBakaXL     LauncherID = "bakaxl"
	LauncherPolymerium LauncherID = "polymerium"
	LauncherXMCL       LauncherID = "xmcl"
)

// Evidence describes how a candidate path was discovered.
type Evidence string

const (
	EvCache          Evidence = "cache"
	EvKnownPaths     Evidence = "known_paths"
	EvRegistry       Evidence = "registry_uninstall"
	EvAppsFolder     Evidence = "shell_appsfolder"
	EvRunningProcess Evidence = "running_process"
	EvStartMenuLnk   Evidence = "start_menu_shortcut"
	EvScan           Evidence = "targeted_scan"
	EvManual         Evidence = "manual_user_choice"
)

// Candidate represents a discovered launcher path entry.
type Candidate struct {
	Launcher   LauncherID `json:"launcher"`
	Path       string     `json:"path"`
	Kind       string     `json:"kind"` // "exe" or "data_dir"
	Evidence   Evidence   `json:"evidence"`
	Confidence string     `json:"confidence"` // "high", "medium", "low"
	UserPicked bool       `json:"user_picked"`
	LastUsed   time.Time  `json:"last_used"`
	LastOK     time.Time  `json:"last_ok"`
	HashHint   string     `json:"hash_hint"`
}

// LauncherCache is the JSON-persisted cache of detected launcher paths.
type LauncherCache struct {
	Version    int                         `json:"version"`
	UpdatedAt  time.Time                   `json:"updated_at"`
	Candidates map[LauncherID][]*Candidate `json:"candidates"`
}

// ── Cache I/O ────────────────────────────────────

// CachePath returns the path where the launcher cache is stored.
func CachePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PolyForge", "launcher_cache.json"), nil
}

// LoadCache reads the cache from disk, returning a fresh cache on error.
func LoadCache() (*LauncherCache, error) {
	path, err := CachePath()
	if err != nil {
		return newCache(), err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return newCache(), nil
		}
		return newCache(), err
	}
	var c LauncherCache
	if err := json.Unmarshal(b, &c); err != nil {
		// Corrupted - start fresh rather than breaking installs.
		return newCache(), nil
	}
	if c.Candidates == nil {
		c.Candidates = map[LauncherID][]*Candidate{}
	}
	return &c, nil
}

// SaveCache writes the cache to disk.
func SaveCache(c *LauncherCache) error {
	path, err := CachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	c.UpdatedAt = time.Now()
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// ── Cache helpers ────────────────────────────────

func newCache() *LauncherCache {
	return &LauncherCache{
		Version:    1,
		UpdatedAt:  time.Now(),
		Candidates: map[LauncherID][]*Candidate{},
	}
}

// UpsertCandidate adds or updates a candidate in the cache.
func UpsertCandidate(cache *LauncherCache, cand *Candidate) {
	if cache.Candidates == nil {
		cache.Candidates = map[LauncherID][]*Candidate{}
	}
	list := cache.Candidates[cand.Launcher]

	for i := range list {
		if strings.EqualFold(list[i].Path, cand.Path) && list[i].Kind == cand.Kind {
			list[i] = mergeCandidate(list[i], cand)
			cache.Candidates[cand.Launcher] = normalizeCandidates(list)
			return
		}
	}
	list = append(list, cand)
	cache.Candidates[cand.Launcher] = normalizeCandidates(list)
}

// BestValidCachedCandidate returns the highest-priority cached candidate
// that passes the given validation function.
func BestValidCachedCandidate(cache *LauncherCache, id LauncherID, validate func(string) error) *Candidate {
	list := cache.Candidates[id]
	if len(list) == 0 {
		return nil
	}

	sort.SliceStable(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.UserPicked != b.UserPicked {
			return a.UserPicked
		}
		if !a.LastOK.Equal(b.LastOK) {
			return a.LastOK.After(b.LastOK)
		}
		return a.LastUsed.After(b.LastUsed)
	})

	for _, cand := range list {
		if err := validate(cand.Path); err == nil {
			cand.LastOK = time.Now()
			return cand
		}
	}
	return nil
}

// PathHint produces a short hash hint for a path (for deduplication).
func PathHint(p string) string {
	h := sha1.Sum([]byte(strings.ToLower(p)))
	return fmt.Sprintf("%x", h[:6])
}

func mergeCandidate(old, newer *Candidate) *Candidate {
	if newer.Confidence != "" {
		old.Confidence = newer.Confidence
	}
	if newer.Evidence != "" {
		old.Evidence = newer.Evidence
	}
	old.UserPicked = old.UserPicked || newer.UserPicked
	if newer.LastUsed.After(old.LastUsed) {
		old.LastUsed = newer.LastUsed
	}
	if newer.LastOK.After(old.LastOK) {
		old.LastOK = newer.LastOK
	}
	if newer.HashHint != "" {
		old.HashHint = newer.HashHint
	}
	return old
}

func normalizeCandidates(list []*Candidate) []*Candidate {
	sort.SliceStable(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.UserPicked != b.UserPicked {
			return a.UserPicked
		}
		if !a.LastOK.Equal(b.LastOK) {
			return a.LastOK.After(b.LastOK)
		}
		return a.LastUsed.After(b.LastUsed)
	})
	if len(list) > 10 {
		list = list[:10]
	}
	return list
}
