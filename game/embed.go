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

		// Get the first line of the file content
		firstLine := string(data)
		if idx := strings.Index(firstLine, "\n"); idx >= 0 {
			firstLine = firstLine[:idx]
		}

		// Skip raw data / bonus files (their first line contains /graw or /bon)
		if strings.Contains(firstLine, "/graw") || strings.Contains(firstLine, "/bon") {
			return nil
		}

		LoadMap = append(LoadMap, data)
		return nil
	})
}
