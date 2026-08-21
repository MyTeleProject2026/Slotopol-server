package game

import (
	"embed"
	"io/fs"
	"path/filepath"
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
		LoadMap = append(LoadMap, data)
		return nil
	})
}
