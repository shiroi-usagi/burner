package burner

import (
	"io"
	"strconv"
)

const (
	ModeNone Mode = iota - 1
	ModeSampleMP4
	ModeFragmentedMP4
	ModeMP4
	ModeTranscode
)

var (
	Modes = []Mode{
		ModeFragmentedMP4,
		ModeMP4,
		ModeTranscode,
		ModeSampleMP4,
	}

	labels = map[Mode]string{
		ModeSampleMP4:     "Sample MP4 (mux)",
		ModeFragmentedMP4: "Fragmented MP4 (HLS)",
		ModeMP4:           "MP4 (mux)",
		ModeTranscode:     "Transcode (softsub)",
	}

	flags = map[string]Mode{
		"smp4":      ModeSampleMP4,
		"fmp4":      ModeFragmentedMP4,
		"mp4":       ModeMP4,
		"transcode": ModeTranscode,
	}
)

type Mode int

// Label is a user friendly representation of m
func (m Mode) Label() string {
	return labels[m]
}

// StringToMode recognises a string representation of
// modes.
//
// If the string is unknown it returns ModeNone.
func StringToMode(s string) Mode {
	for k, m := range flags {
		if k == s {
			return m
		}
	}
	return ModeNone
}

// ReadMode reads a rune from the given reader which
// value will be used to determine a Mode.
//
// If the rune is unknown it returns ModeNone.
func ReadMode(reader io.RuneReader) Mode {
	r, _, _ := reader.ReadRune()
	i, err := strconv.Atoi(string(r))
	if err != nil {
		return ModeNone
	}
	for _, m := range Modes {
		if int(m) == i {
			return m
		}
	}
	return ModeNone
}
