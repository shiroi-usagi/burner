package ffmpeg

import (
	"fmt"
	"github.com/shiroi-usagi/burner/ziputil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	releaseRawurl = "https://github.com/GyanD/codexffmpeg/releases/download/7.1.1/ffmpeg-7.1.1-essentials_build.zip"
	releaseFolder = strings.TrimSuffix(path.Base(releaseRawurl), ".zip")
)

// ExecutableFallback downloads ffmpeg binaries for Windows from a trusted source.
// The binaries are for ffmpeg, ffplay, and ffrpobe.
func ExecutableFallback(file string) (string, error) {
	if !knownBinary(file) {
		return "", ErrUnknownBinary
	}
	root, err := os.Executable()
	if err != nil {
		root = "."
	} else {
		root = filepath.Dir(root)
	}
	binDir, _ := filepath.Abs(filepath.Join(root, "bin"))
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
	if _, err := ziputil.UnzipAll(releaseRawurl, filepath.Join(root, "bin", "ffmpeg")); err != nil {
		return "", fmt.Errorf("could not unzip executable: %w", err)
	}
	return executable, nil
}
