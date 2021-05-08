package filepathutil

import (
	"os"
	"path/filepath"
)

// ListFilesWithExt lists all the files with the given extension in the given directory.
func ListFilesWithExt(directory string, ext ...string) []string {
	extM := map[string]bool{}
	for _, e := range ext {
		extM[e] = true
	}

	var files []string
	_ = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() || !extM[filepath.Ext(path)] {
			return nil
		}

		files = append(files, path)

		return nil
	})
	return files
}
