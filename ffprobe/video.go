package ffprobe

import (
	"encoding/json"
	"os/exec"
	"strconv"
)

type format struct {
	Duration string `json:"duration"`
}

type entries struct {
	Format format `json:"format"`
}

func Duration(path, input string) (float64, error) {
	var args []string
	args = append(args, "-i", input)                        // Input file url
	args = append(args, "-show_entries", "format=duration") // Set list of entries to show.
	args = append(args, "-v", "quiet")                      // Show nothing at all; be silent.
	args = append(args, "-of", "json")                      // Set the output printing format.
	cmd := exec.Command(path, args...)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var e entries
	if err = json.Unmarshal(out, &e); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(string(e.Format.Duration), 64)
}
