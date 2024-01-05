package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

//go:embed data/*
var data embed.FS

func main() {
	var templates TemplateJSON

	// Walk through the directory
	err := fs.WalkDir(data, "data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		dir, file := filepath.Split(path)
		dir = filepath.Base(dir) // get the category name

		// Check for P or LS in filename
		name := strings.TrimSuffix(file, filepath.Ext(file))
		landscape := strings.HasPrefix(name, "LS ")
		name = strings.TrimPrefix(name, "P ")
		name = strings.TrimPrefix(name, "LS ")
		name = strings.TrimSpace(name)
		iconcode := DefaultIconPortrait
		if landscape {
			iconcode = DefaultIconLandscape
		}

		templates.Templates = append(templates.Templates, struct {
			Name       string   `json:"name"`
			Filename   string   `json:"filename"`
			IconCode   string   `json:"iconCode"`
			Categories []string `json:"categories"`
			Landscape  bool     `json:"landscape,omitempty"`
		}{
			Name:       name,
			IconCode:   iconcode,
			Filename:   path,
			Categories: []string{dir},
			Landscape:  landscape,
		})

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking through embedded files: %v", err)
	}

	// Print the JSON for demonstration purposes
	jsonData, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling data to JSON: %v", err)
	}
	log.Println(string(jsonData))
}
