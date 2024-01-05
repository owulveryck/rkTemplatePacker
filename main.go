package main

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"os"
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
		// Remove the "data/" prefix from the path
		path = strings.TrimPrefix(path, "data/")
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
	targetDir := "/tmp/output" // Define your target directory here

	err = copyEmbeddedFiles(targetDir)
	if err != nil {
		log.Fatalf("Error copying embedded files: %v", err)
	}

	log.Println("Files copied successfully to", targetDir)

}

// copyEmbeddedFiles copies files from embedded file system to a target directory on disk.
func copyEmbeddedFiles(targetDir string) error {

	return fs.WalkDir(data, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == "." {
			return nil
		}

		// Construct the target path
		subpath := strings.TrimPrefix(path, "data/")
		targetPath := filepath.Join(targetDir, subpath)

		// If it's a directory, create it
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Open the source file
		srcFile, err := data.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Create the target file
		targetFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		// Copy the contents
		_, err = io.Copy(targetFile, srcFile)
		return err
	})
}
