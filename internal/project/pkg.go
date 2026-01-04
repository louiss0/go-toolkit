package project

import (
	"errors"
	"os"
	"path/filepath"
)

const mainTemplate = `package main

func main() {}
`

type Options struct {
	WriteMain     bool
	WriteInternal bool
}

func EnsureLayout(root string, options Options) error {
	if root == "" {
		return errors.New("root path is required")
	}

	if options.WriteInternal {
		if err := ensureInternalDir(root); err != nil {
			return err
		}
	}

	if options.WriteMain {
		if err := ensureMainFile(root); err != nil {
			return err
		}
	}

	return nil
}

func ensureInternalDir(root string) error {
	return os.MkdirAll(filepath.Join(root, "internal"), 0o755)
}

func ensureMainFile(root string) error {
	path := filepath.Join(root, "main.go")
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(mainTemplate)
	return err
}
