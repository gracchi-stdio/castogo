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
	viteManifestOnce sync.Once
	viteManifestData map[string]viteManifestEntry
	viteManifestErr  error
)

func loadViteManifest() {
	manifestPath := filepath.Join("public", ".vite", "manifest.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		viteManifestErr = err
		return
	}

	data := map[string]viteManifestEntry{}
	if err := json.Unmarshal(content, &data); err != nil {
		viteManifestErr = err
		return
	}

	viteManifestData = data
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

func viteAssets(entry string) ([]string, string) {
	viteManifestOnce.Do(loadViteManifest)
	if viteManifestErr != nil {
		fmt.Printf("warn: vite manifest not found (%v); run 'just frontend-build'\n", viteManifestErr)
		return nil, ""
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
