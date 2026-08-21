package game

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed slot/*/*/*.yaml
var embeddedYAML embed.FS

func init() {
	_ = fs.WalkDir(embeddedYAML, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" {
			return nil
		}
		data, _ := embeddedYAML.ReadFile(path)
		content := string(data)

		// Find the first non‑empty, non‑comment line
		var firstLine string
		for _, line := range strings.Split(content, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			firstLine = trimmed
			break
		}

		// Only keep reel definition files (first line ends with "/rmap")
		if !strings.HasSuffix(firstLine, "/rmap") {
			return nil
		}

		LoadMap = append(LoadMap, data)
		return nil
	})
}
