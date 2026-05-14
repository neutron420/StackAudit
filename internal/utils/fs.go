package utils

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type WalkOptions struct {
	MaxSize    int64
	Extensions map[string]bool
	SkipDirs   map[string]bool
	Ignore     IgnoreMatcher
}

func WalkFiles(root string, opts WalkOptions) ([]string, error) {
	var files []string
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if opts.Ignore != nil && opts.Ignore(path, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		if d.IsDir() {
			if opts.SkipDirs != nil && opts.SkipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		if opts.Extensions != nil {
			ext := strings.ToLower(filepath.Ext(name))
			if ext == "" && opts.Extensions["*"] {
				// allow extensionless files
			} else if !opts.Extensions[ext] {
				return nil
			}
		}

		if opts.MaxSize > 0 {
			info, statErr := d.Info()
			if statErr != nil {
				return statErr
			}
			if info.Size() > opts.MaxSize {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, walkErr
}
