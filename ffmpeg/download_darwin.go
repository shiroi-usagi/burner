package ffmpeg

import (
	"fmt"
	"github.com/shiroi-usagi/burner/ziputil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	releaseRawurl = "https://evermeet.cx/ffmpeg/%s-7.1.1.zip"
)

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
	executable := filepath.Join(binDir, "ffmpeg", file)
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
	if _, err := ziputil.UnzipAll(fmt.Sprintf(releaseRawurl, file), filepath.Join(root, "bin", "ffmpeg")); err != nil {
		return "", fmt.Errorf("could not unzip executable: %w", err)
	}
	return executable, nil
}
