package burn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/shiroi-usagi/burner"
	"github.com/shiroi-usagi/burner/ffmpeg"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
)

var Cmd = &cobra.Command{
	Use:   "burn",
	Short: "transcode video",
	Long: `Transcode all files in the input folder and save
the result to the output folder.`,
}

func init() {
	Cmd.Run = run // break init cycle
}

var (
	verbose = Cmd.Flags().BoolP("verbose", "v", false, "make output verbose")

	mode = Cmd.Flags().StringP("mode", "m", "", `mode of the encoding
  smp4 - Sample MP4. Encodes a sample with the subtitle burned on the video. Creates hardsub.
  fmp4 - Fragmented MP4. Encodes a fragmented video (HLS) with the subtitle burned on the video. Creates hardsub.
  mp4 - MP4. Encodes a video with the subtitle burned on the video. Creates hardsub.
  transcode - Transcode. Encodes a video with the given options while keeping the original settings. Creates softsub.`)

	inputDir  = Cmd.Flags().StringP("input", "i", "./in", "directory of the input files")
	outputDir = Cmd.Flags().StringP("output", "o", "./out", "directory of the output files")

	ignoreFontError = Cmd.Flags().Bool("ignore-font-error", false, "skip font errors during encode")

	videoHeight      = Cmd.Flags().Int("v-height", burner.DefaultHeight, "target video height")
	videoBitrate     = Cmd.Flags().String("v-bitrate", burner.DefaultBitrate, "target video bitrate")
	videoKeepBitrate = Cmd.Flags().Bool("v-keep-bitrate", false, "disables bitrate modification when the original file size smaller than the expected")
	videoUpscaling   = Cmd.Flags().Bool("v-upscaling", false, "enable/disable upscaling")
)

func run(_ *cobra.Command, args []string) {
	absIn, err := filepath.Abs(*inputDir)
	if err != nil {
		fmt.Println("Could not create absolute representation of input folder")
	}
	absOut, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Println("Could not create absolute representation of output folder")
	}
	ffmpegExecutable, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println("ffmpeg is not found in path, will try fallback")
		ffmpegExecutable, err = ffmpeg.ExecutableFallback("ffmpeg")
		if err != nil {
			fmt.Println(err)
		}
	}
	ffprobeExecutable, err := exec.LookPath("ffprobe")
	if err != nil {
		fmt.Println("ffprobe is not found in path, will try fallback")
		ffprobeExecutable, err = ffmpeg.ExecutableFallback("ffprobe")
		if err != nil {
			fmt.Println(err)
		}
	}
	currentV := burner.GetVersionInfo().Version
	fmt.Println(fmt.Sprintf("Current version: `%s`", currentV))
	latestV, err := latestVersion()
	if err != nil {
		fmt.Println("Latest version: Unknown")
	} else {
		fmt.Println(fmt.Sprintf("Latest version: `%s`", latestV))
		if devVersion != currentV && semver.Compare(latestV, currentV) > 0 {
			fmt.Println("Consider upgrading your version")
		}
	}

	reader := bufio.NewReader(os.Stdin)
	selectedMode := burner.StringToMode(*mode)
	for selectedMode == burner.ModeNone {
		fmt.Println("Select mode:")
		for _, m := range burner.Modes {
			fmt.Println(fmt.Sprintf("[%d] %s", m, m.Label()))
		}
		selectedMode = burner.ReadMode(reader)
	}

	burner.Burn(burner.Config{
		Verbose: *verbose,

		Mode: selectedMode,

		InputDir:    absIn,
		OutputDir:   absOut,
		FFmpegPath:  ffmpegExecutable,
		FFprobePath: ffprobeExecutable,

		IgnoreFontError: *ignoreFontError,

		Video: burner.VideoConf{
			Height:      *videoHeight,
			Bitrate:     *videoBitrate,
			KeepBitrate: *videoKeepBitrate,
			Upscaling:   *videoUpscaling,
		},
	})
}

type releasePayload struct {
	TagName string `json:"tag_name"`
}

var devVersion = "v0.0.0-SNAPSHOT"

func latestVersion() (string, error) {
	c := http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get("https://api.github.com/repos/shiroi-usagi/burner/releases")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var payload []releasePayload
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return "", err
	}
	if len(payload) == 0 {
		return "v0.0.0-SNAPSHOT", nil
	}
	sort.Slice(payload, func(i, j int) bool {
		return semver.Compare(payload[i].TagName, payload[j].TagName) > 0
	})
	return payload[0].TagName, nil
}
