package file

import (
	"os"
	"path/filepath"
	"sort"
)

func GetFileList(filename string) []string {
	files, err := filepath.Glob(filename)
	if err != nil {
		return []string{}
	}

	return files
}

func GetFileListSorted(filename string) []string {
	files := GetFileList(filename)

	sort.Strings(files)

	return files
}

func Remove(filename string) (string, error) {
	if err := os.Remove(filename); err != nil {
		return "", err
	}
	return filename, nil
}

func Removes(files []string) []string {
	removedFiles := []string{}
	for _, f := range files {
		if _, err := Remove(f); err != nil {
			continue
		}
		removedFiles = append(removedFiles, f)
	}

	return removedFiles
}

func RemoveByPattern(pattern string) []string {
	files := GetFileList(pattern)

	return Removes(files)
}
