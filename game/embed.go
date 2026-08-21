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
		// Skip raw data and bonus directories
		if strings.Contains(path, "/graw/") || strings.Contains(path, "/bon/") {
			return nil
		}
		// Skip non-YAML files
		if filepath.Ext(path) != ".yaml" {
			return nil
		}
		data, _ := embeddedYAML.ReadFile(path)
		LoadMap = append(LoadMap, data)
		return nil
	})
}
