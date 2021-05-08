// +build !windows

package ffmpeg

import "errors"

func ExecutableFallback(file string) (string, error) {
	return "", errors.New("no fallback provided")
}
