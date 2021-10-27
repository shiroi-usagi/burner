package ziputil

import (
	"archive/zip"
	"fmt"
	"github.com/jeffallen/seekinghttp"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnzipAll will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func UnzipAll(src string, dest string) ([]string, error) {
	var filenames []string

	var zr *zip.Reader
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		body := seekinghttp.New(src)
		size, err := body.Size()
		if err != nil {
			return nil, err
		}
		r, err := zip.NewReader(body, size)
		if err != nil {
			return nil, err
		}
		zr = r
	} else {
		f, err := os.Open(src)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}
		r, err := zip.NewReader(f, fi.Size())
		if err != nil {
			return nil, err
		}
		zr = r
	}

	for _, f := range zr.File {
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return filenames, err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
