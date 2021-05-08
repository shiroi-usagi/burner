package filepathutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestListFilesWithExt(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestListFilesWithExt")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("finding working dir:", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatal("entering temp dir:", err)
	}
	defer os.Chdir(origDir)

	files := []string{
		"test.mp4",
		"test.mkv",
		"test",
		"test.txt",
	}

	for _, f := range files {
		_, err := os.Create(f)
		if err != nil {
			t.Fatal(err)
		}
	}

	type args struct {
		directory string
		ext       []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty extension list",
			args: args{directory: tmpDir},
		},
		{
			name: "one extension",
			args: args{directory: tmpDir, ext: []string{".mkv"}},
			want: []string{filepath.Join(tmpDir, "test.mkv")},
		},
		{
			name: "multiple extension",
			args: args{directory: tmpDir, ext: []string{".mkv", ".mp4"}},
			want: []string{filepath.Join(tmpDir, "test.mkv"), filepath.Join(tmpDir, "test.mp4")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ListFilesWithExt(tt.args.directory, tt.args.ext...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListFilesWithExt() = %v, want %v", got, tt.want)
			}
		})
	}
}
