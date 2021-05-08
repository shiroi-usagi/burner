package ffmpeg

import "errors"

var (
	ErrUnknownBinary = errors.New("unknown binary name provided")
)

var (
	knownBinaries = []string{"ffmpeg", "ffplay", "ffprobe"}
)
