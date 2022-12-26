package file

import (
	"errors"
	"os"
	"sort"
	"testing"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func createTempFiles(number int, pattern string) ([]string, error) {
	filenames := []string{}
	for i := 0; i < number; i++ {
		f, err := os.CreateTemp("", pattern)
		if err != nil {
			return nil, errors.New("Cannot create temporary file.")
		}
		defer f.Close()

		filenames = append(filenames, f.Name())
	}

	return filenames, nil
}

func removeTempFiles(files []string) {
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			continue
		}
	}
}

func TestFileNameMatchOnlyOne(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Errorf("Cannot create temporary file.")
	}
	defer f.Close()
	defer os.Remove(f.Name())

	got := GetFileList(f.Name())
	want := f.Name()

	if len(got) != 1 || ! contains(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestFileNameMatchThreeFiles(t *testing.T) {
	tempFiles, err := createTempFiles(3, "test-*.json")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer removeTempFiles(tempFiles)

	got := GetFileList("/tmp/test-*")
	want := tempFiles

	if len(got) != 3 {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestFileNameMatchThreeFilesSortedAscending(t *testing.T) {
	tempFiles, err := createTempFiles(3, "test-*.json")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer removeTempFiles(tempFiles)

	got := GetFileList("/tmp/test-*")
	sort.Strings(got)

	if ! sort.StringsAreSorted(got) {
		t.Errorf("%q is not sorted", got)
	}
}

func TestFileNameMatchSortedAscendingGet5FirstValue(t *testing.T) {
	tempFiles, err := createTempFiles(10, "test-*.json")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer removeTempFiles(tempFiles)

	got := GetFileList("/tmp/test-*")
	sort.Strings(got)

	removedFiles := got[:5]

	if len(removedFiles) != 5 || ! sort.StringsAreSorted(removedFiles) {
		t.Errorf("length of %q, is not 5. All data: %q", removedFiles, got)
	}
}

func TestRemoveOneFile(t *testing.T) {
	tempFiles, err := createTempFiles(1, "test-*.json")
	if err != nil {
		t.Errorf(err.Error())
	}

	Remove(tempFiles[0])

	got := GetFileList(tempFiles[0])

	if len(got) > 0 {
		t.Errorf("got %q, wanted []", got)
		removeTempFiles(tempFiles)
	}
}

func TestRemoveFiveFile(t *testing.T) {
	tempFiles, err := createTempFiles(5, "test-*.json")
	if err != nil {
		t.Errorf(err.Error())
	}

	pattern := "/tmp/test-*"

	removed := RemoveByPattern(pattern)

	got := GetFileList(pattern)

	if len(got) > 0 && len(removed) == 5 {
		t.Errorf("got %q, wanted []", got)
		removeTempFiles(tempFiles)
	}
}
