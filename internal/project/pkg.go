package project

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/louiss0/go-toolkit/validation"
)

const mainTemplate = `package main

func main() {}
`

type Options struct {
	WriteMain     bool
	WriteInternal bool
}

func EnsureLayout(root string, options Options) error {
	if _, err := validation.RequiredString(root, "root path"); err != nil {
		return err
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

func ensureMainFile(root string) (err error) {
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
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	_, err = file.WriteString(mainTemplate)
	return err
}
