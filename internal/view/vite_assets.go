package view

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type viteManifestEntry struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

var (
	viteManifestMu     sync.RWMutex
	viteManifestData   map[string]viteManifestEntry
	viteManifestLoaded bool
)

func loadViteManifest() error {
	manifestPath := filepath.Join("public", ".vite", "manifest.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	data := map[string]viteManifestEntry{}
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}

	viteManifestData = data

	viteManifestLoaded = true
	return nil
}

func normalizeAssetPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "/") {
		return trimmed
	}

	return "/" + trimmed
}

func ViteAssets(entry string) ([]string, string) {
	viteManifestMu.RLock()
	loaded := viteManifestLoaded
	viteManifestMu.RUnlock()

	if !loaded {
		viteManifestMu.Lock()
		if !viteManifestLoaded {
			if err := loadViteManifest(); err != nil {
				fmt.Printf("error loading vite manifest: %v\n", err)
				viteManifestMu.Unlock()
				return nil, ""
			}
		}
		viteManifestMu.Unlock()
	}

	manifestEntry, ok := viteManifestData[entry]
	if !ok {
		fmt.Printf("warn: entry %q not found in vite manifest\n", entry)
		return nil, ""
	}

	css := make([]string, 0, len(manifestEntry.CSS))
	for _, cssPath := range manifestEntry.CSS {
		normalized := normalizeAssetPath(cssPath)
		if normalized != "" {
			css = append(css, normalized)
		}
	}

	return css, normalizeAssetPath(manifestEntry.File)
}
