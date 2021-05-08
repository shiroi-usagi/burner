package burn

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/shiroi-usagi/burner"
	"github.com/shiroi-usagi/burner/ffmpeg"
	"github.com/shiroi-usagi/pkg/command"
	"golang.org/x/mod/semver"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
)

var Cmd = &command.Subcommand{
	Name:      "burn",
	Arguments: "[-v] [-mode m] [-input dir] [-output dir] [-ffmpeg path] [-v-height num] [-v-bitrate num] [-v-upscaling num]",
	Short:     "transcode video",
	Long: `Transcode all files in the input folder and save
the result to the output folder.
`,
	Flag: flag.NewFlagSet("", flag.ExitOnError),
}

func init() {
	Cmd.Run = run // break init cycle

	var buf bytes.Buffer
	Cmd.Flag.SetOutput(&buf)
	Cmd.Flag.PrintDefaults()
	Cmd.Long += "\n"
	Cmd.Long += buf.String()
}

var (
	verbose = Cmd.Flag.Bool("v", false, "make output verbose")

	mode = Cmd.Flag.String("mode", "", `mode of the encoding

Possible values:
 - smp4
   Stands for Sample MP4. Encodes a sample with the subtitle burned on the video. Creates hardsub.
 - fmp4
   Stands for Fragmented MP4. Encodes a fragmented video (HLS) with the subtitle burned on the video. Creates hardsub.
 - mp4
   Stands for MP4. Encodes a video with the subtitle burned on the video. Creates hardsub.
 - transcode
   Stands for Transcode. Encodes a video with the given options while keeping the original settings. Creates softsub.`)

	inputDir  = Cmd.Flag.String("input", "./in", "directory of the input files")
	outputDir = Cmd.Flag.String("output", "./out", "directory of the output files")

	videoHeight      = Cmd.Flag.Int("v-height", 720, "target video height")
	videoBitrate     = Cmd.Flag.String("v-bitrate", "1371k", "target video bitrate")
	videoKeepBitrate = Cmd.Flag.Bool("v-keep-bitrate", false, "disables bitrate modification when the original file size smaller than the expected")
	videoUpscaling   = Cmd.Flag.Bool("v-upscaling", false, "enable/disable upscaling")
)

func run(_ *command.Subcommand, _ []string) {
	absIn, err := filepath.Abs(*inputDir)
	if err != nil {
		log.Fatal("Could not create absolute representation of input folder")
	}
	absOut, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatal("Could not create absolute representation of output folder")
	}
	ffmpegExecutable, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Println("ffmpeg is not found in path")
		ffmpegExecutable, err = ffmpeg.ExecutableFallback("ffmpeg")
		if err != nil {
			log.Fatal(err)
		}
	}
	ffprobeExecutable, err := exec.LookPath("ffprobe")
	if err != nil {
		log.Println("ffprobe is not found in path")
		ffprobeExecutable, err = ffmpeg.ExecutableFallback("ffprobe")
		if err != nil {
			log.Fatal(err)
		}
	}
	currentV := burner.GetVersionInfo().Version
	fmt.Println(fmt.Sprintf("Current version: `%s`", currentV))
	latestV, err := latestVersion()
	if err != nil {
		log.Print("Latest version: Unknown")
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
