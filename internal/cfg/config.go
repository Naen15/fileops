package cfg

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
)

// Config regroupe toutes les clés possibles.
type Config struct {
	DefaultFile string `json:"default_file"`
	BaseDir     string `json:"base_dir"`
	OutDir      string `json:"out_dir"`
	DefaultExt  string `json:"default_ext"`
	WikiLang    string `json:"wiki_lang"`
	ProcessTopN int    `json:"process_top_n"`
}

func Load() (Config, error) {
	// 1) valeurs par défaut
	cfg := Config{
		DefaultFile: "data/input.txt",
		BaseDir:     "data",
		OutDir:      "out",
		DefaultExt:  ".txt",
		WikiLang:    "fr",
		ProcessTopN: 10,
	}

	path := flag.Lookup("config").Value.String()
	if path == "" {
		path = "config.json"
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	_ = json.Unmarshal(b, &cfg)

	if !filepath.IsAbs(cfg.OutDir) {
		cfg.OutDir = filepath.Clean(filepath.Join(filepath.Dir(path), cfg.OutDir))
	}
	return cfg, nil
}
