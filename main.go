package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TemplateGlobalDir      string `envconfig:"TEMPLATE_GLOBAL_DIR" default:"/usr/share/remarkable/templates"`
	TemplateJSONFile       string `envconfig:"TEMPLATE_JSON_FILE" default:"/usr/share/remarkable/templates/templates.json"`
	TemplateJSONBackupFile string `envconfig:"TEMPLATE_JSON_BACKUP_FILE" default:"/home/root/templates_backup.json"`
}

var (
	DefaultIconPortrait  = "\ue9fe"
	DefaultIconLandscape = "\ue9fd"
)

//go:embed data/*
var data embed.FS
var cfg Config

func main() {
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	originalContent, err := os.ReadFile(cfg.TemplateJSONFile)
	if err != nil {
		log.Fatal(err)
	}

	var template TemplateJSON
	err = UnmarshalStrict(originalContent, &template)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	templates, err := generateTemplatesFromEmbededFiles()

	if err != nil {
		log.Fatalf("Error walking through embedded files: %v", err)
	}

	// Add the template to the original if it does not exists
	for _, customTemplate := range templates {
		exists := false
		for _, originalTemplate := range template.Templates {
			if customTemplate.Name == originalTemplate.Name {
				exists = true
			}
		}
		if !exists {
			template.Templates = append(template.Templates, customTemplate)
		}
	}

	err = copyEmbededFiles(cfg.TemplateGlobalDir)
	if err != nil {
		log.Fatalf("Error copying embedded files: %v", err)
	}

	log.Println("Files copied successfully to", cfg.TemplateGlobalDir)

	err = writeJSONTemplate(template)
	if err != nil {
		log.Fatalf("Error copying embedded files: %v", err)
	}
	fmt.Printf("you can compare with: \ndiff <(jq --sort-keys . %v) <(jq --sort-keys . %v)\n", cfg.TemplateJSONBackupFile, cfg.TemplateJSONFile)

}

func writeJSONTemplate(tmpl TemplateJSON) error {
	// Create a backup of the existing file if it exists
	if _, err := os.Stat(cfg.TemplateJSONFile); err == nil {
		err := copyFile(cfg.TemplateJSONFile, cfg.TemplateJSONBackupFile)
		if err != nil {
			return err
		}
	}

	// Encode data to JSON
	var jsonData bytes.Buffer

	enc := json.NewEncoder(&jsonData)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "   ")
	err := enc.Encode(tmpl)
	if err != nil {
		return err
	}

	// Write JSON data to file
	err = ioutil.WriteFile(cfg.TemplateJSONFile, jsonData.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func generateTemplatesFromEmbededFiles() ([]Template, error) {
	var templates = make([]Template, 0)
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

		templates = append(templates, Template{
			Name:       name,
			IconCode:   iconcode,
			Filename:   path,
			Categories: []string{dir},
			Landscape:  landscape,
		})

		return nil
	})
	return templates, err
}

// copyEmbededFiles copies files from embedded file system to a target directory on disk.
func copyEmbededFiles(targetDir string) error {

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

// UnmarshalStrict enforces strict unmarshaling
func UnmarshalStrict(data []byte, v interface{}) error {
	dec := json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data)))
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return err
	}

	// Check for extraneous data
	if dec.More() {
		return fmt.Errorf("extra data in JSON")
	}

	return nil
}

// copyFile copies the contents of the file named src to the file named dst.
// The file will be created if it does not already exist. If the destination file exists,
// all it's contents will be replaced by the contents of the source file.
func copyFile(src, dst string) error {
	// Open the source file for reading
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file for writing
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(destFile, sourceFile)
	return err
}
