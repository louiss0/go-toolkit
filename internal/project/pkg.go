package project

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/louiss0/go-toolkit/validation"
)

//go:embed assets/templates
var templateFiles embed.FS

const (
	TemplateAPI = "api"
	TemplateCLI = "cli"
	TemplateLib = "lib"
)

type Options struct {
	Template string
}

func EnsureLayout(root string, options Options) error {
	if _, err := validation.RequiredString(root, "root path"); err != nil {
		return err
	}

	template, err := resolveTemplate(options.Template)
	if err != nil {
		return err
	}

	templateTree, err := fs.Sub(templateFiles, templatePath(template))
	if err != nil {
		return err
	}

	return writeTemplate(root, templateTree)
}

func TemplateValues() []string {
	return []string{TemplateAPI, TemplateCLI, TemplateLib}
}

func resolveTemplate(template string) (string, error) {
	if template == "" {
		return TemplateAPI, nil
	}

	for _, allowedTemplate := range TemplateValues() {
		if template == allowedTemplate {
			return template, nil
		}
	}

	return "", fmt.Errorf("template must be one of: %v", TemplateValues())
}

func templatePath(template string) string {
	return fmt.Sprintf("assets/templates/%s", template)
}

func writeTemplate(root string, templateTree fs.FS) error {
	return fs.WalkDir(templateTree, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		targetPath := filepath.Join(root, materializePath(path))
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		content, err := fs.ReadFile(templateTree, path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, content, 0o644)
	})
}

func materializePath(path string) string {
	return strings.TrimSuffix(path, ".tmpl")
}
