package ffmpeg

import "errors"

var (
	ErrUnknownBinary = errors.New("unknown binary name provided")
)

var (
	knownBinaries = []string{"ffmpeg", "ffplay", "ffprobe"}
)

func knownBinary(file string) bool {
	for _, binary := range knownBinaries {
		if binary == file {
			return true
		}
	}
	return false
}
