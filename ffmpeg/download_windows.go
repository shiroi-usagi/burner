package ffmpeg

import (
	"fmt"
	"github.com/shiroi-usagi/burner/ziputil"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	releaseFolder = "ffmpeg-4.3.1-2020-11-19-essentials_build"
	releaseRawurl = "https://github.com/GyanD/codexffmpeg/releases/download/4.3.1-2020-11-19/ffmpeg-4.3.1-2020-11-19-essentials_build.zip"
)

func knownBinary(file string) bool {
	for _, binary := range knownBinaries {
		if binary == file {
			return true
		}
	}
	return false
}

// ExecutableFallback downloads ffmpeg binaries for Windows from a trusted source.
// The binaries are for ffmpeg, ffplay, and ffrpobe.
func ExecutableFallback(file string) (string, error) {
	if !knownBinary(file) {
		return "", ErrUnknownBinary
	}
	binDir, _ := filepath.Abs(filepath.Join(".", "bin"))
	executable := filepath.Join(binDir, "ffmpeg", releaseFolder, "bin", file+".exe")
	if _, err := os.Stat(executable); !os.IsNotExist(err) {
		// already downloaded
		return executable, nil
	}
	log.Printf("Downloading %s", file)
	t := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	defer close(done)
	fmt.Print("=")
	defer fmt.Println()
	go func() {
		for {
			select {
			case <-t.C:
				fmt.Print("=")
			case <-done:
				t.Stop()
				return
			}
		}
	}()
	zippath, err := downloadFile(binDir, releaseRawurl)
	if err != nil {
		return "", fmt.Errorf("could not download release: %w", err)
	}
	if _, err := ziputil.UnzipAll(zippath, "./bin/ffmpeg"); err != nil {
		return "", fmt.Errorf("could not unzip executable: %w", err)
	}
	os.Remove(zippath)
	return executable, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(targetDir string, rawurl string) (string, error) {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}
	c := http.Client{Timeout: 5 * time.Minute}
	resp, err := c.Get(rawurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var filename string
	if _, ok := resp.Header["Content-Disposition"]; ok {
		cd := resp.Header.Get("Content-Disposition")
		t, param, err := mime.ParseMediaType(cd)
		if err != nil {
			return "", err
		}
		if t == "attachment" {
			filename = param["filename"]
		}
	}
	if filename == "" {
		filename = path.Base(resp.Request.URL.Path)
	}
	fpath := filepath.Join(targetDir, filename)
	out, err := os.Create(fpath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return filepath.Abs(fpath)
}
