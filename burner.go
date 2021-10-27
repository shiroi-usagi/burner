package burner

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/shiroi-usagi/burner/commandline"
	"github.com/shiroi-usagi/burner/ffmpeg"
	"github.com/shiroi-usagi/burner/ffprobe"
	"github.com/shiroi-usagi/burner/filepathutil"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	DefaultHeight  = 720
	DefaultBitrate = "1371k"

	// supportedInputExt filters the files from the input directory
	supportedInputExt = []string{".mkv", ".mp4", ".avs"}
)

type VideoConf struct {
	Height      int
	Bitrate     string
	Upscaling   bool
	KeepBitrate bool
}

type Config struct {
	Verbose bool

	Mode Mode

	InputDir    string
	OutputDir   string
	FFmpegPath  string
	FFprobePath string

	Video VideoConf

	IgnoreFontError bool
}

func Burn(conf Config) {
	if _, err := os.Stat(conf.InputDir); err != nil {
		log.Fatal("missing input directory")
	}
	if _, err := os.Stat(conf.OutputDir); err != nil {
		log.Fatal("missing output directory")
	}

	var factory factoryFunc
	switch conf.Mode {
	case ModeSampleMP4:
		factory = ffmpeg.NewSampleMp4Transcoder
	case ModeFragmentedMP4:
		factory = ffmpeg.NewFragmentedMp4Transcoder
	case ModeMP4:
		factory = ffmpeg.NewMp4Transcoder
	case ModeTranscode:
		factory = ffmpeg.NewTranscoder
	default:
		log.Print("Was not able to detect mode")
		return
	}

	if conf.Verbose {
		log.Print(conf.FFmpegPath)
	}

	files := filepathutil.ListFilesWithExt(conf.InputDir, supportedInputExt...)
	l := len(files)

	cmdOut := &modifiableOutput{Stdout: os.Stdout}
	for i, file := range files {
		log.Printf("[%03d/%03d] %s", i+1, l, filepath.Base(file))
		if err := burn(cmdOut, file, factory, conf); err != nil {
			log.Print(err)
			continue
		}
	}
}

type factoryFunc func(executable string, input string, outDir string, bitrate string, f ffmpeg.Filter) *ffmpeg.Transcoder

func burn(cmdOut *modifiableOutput, file string, factory factoryFunc, conf Config) error {
	// Avoid dealing with escaping characters in complex filter
	slink := filepath.Join(conf.OutputDir, "tmp"+filepath.Ext(file))
	_ = os.Remove(slink)
	err := os.Link(file, slink)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(slink)
	}()

	// For YUV 4:2:0 chroma subsampled outputs width and height has to be divisible by 2
	f := ffmpeg.Filter{Subtitle: slink, Width: -2, Height: conf.Video.Height, Upscaling: conf.Video.Upscaling}

	if !conf.Video.KeepBitrate && conf.FFprobePath != "" {
		duration, err := ffprobe.Duration(conf.FFprobePath, file)
		if err != nil {
			return err
		}

		expectedSize := calcExpectedSize(duration, ffmpeg.BitrateToKilobit(conf.Video.Bitrate))
		stat, _ := os.Stat(file)
		size := float64(stat.Size())
		if size < expectedSize {
			kilobit := (size * 8 / 1024 / duration) - 128
			conf.Video.Bitrate = ffmpeg.KilobitToBitrate(int64(kilobit))
			log.Printf("bitrate was modified to %s", conf.Video.Bitrate)
		}
	}

	t := factory(conf.FFmpegPath, file, conf.OutputDir, conf.Video.Bitrate, f)

	if err := os.MkdirAll(t.OutDir(), 0755); err != nil {
		return err
	}

	defer func() {
		// Remove FFmpeg logs
		_ = os.Remove(filepath.Join(t.OutDir(), "ffmpeg2pass-0.log"))
		_ = os.Remove(filepath.Join(t.OutDir(), "ffmpeg2pass-0.log.mbtree"))
	}()
	if err := runCommand(cmdOut, t.FirstPass(), conf); err != nil {
		return err
	}

	if err := runCommand(cmdOut, t.SecondPass(), conf); err != nil {
		return err
	}

	return nil
}

func calcExpectedSize(duration float64, bitrate int64) float64 {
	return (float64(bitrate) + 128) * duration / 8 * 1024
}

// runCommand runs the given command while writing the output to console.
//
// The verbose argument makes the output more talkative.
func runCommand(out io.Writer, cmd *exec.Cmd, conf Config) error {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if conf.Verbose {
		fmt.Println(cmd)
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		s := bufio.NewScanner(stdout)
		var sb strings.Builder
		var i uint8
		for s.Scan() {
			line := s.Text()
			if strings.HasPrefix(line, "progress=") {
				// stats_period flag not available in all versions
				if i == 0 {
					fmt.Println(sb.String())
				}
				i = (i + 1) % 4 // default stats_period is 0.5 seconds, we only need info every 2 seconds
				sb.Reset()
				// The last key of a sequence of progress information is always "progress".
				continue
			}
			if conf.Verbose {
				sb.WriteString(line)
				sb.WriteString(" ")
			} else {
				switch {
				default:
					sb.WriteString(line)
					sb.WriteString(" ")
				case strings.HasPrefix(line, "stream_"),
					strings.HasPrefix(line, "out_time_us="), strings.HasPrefix(line, "out_time_ms="),
					strings.HasPrefix(line, "dup_frames="), strings.HasPrefix(line, "drop_frames="):
					// ignore
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		s := bufio.NewScanner(stderr)
		h := ffmpeg.Printer()
		if !conf.IgnoreFontError {
			h = ffmpeg.KillOnReplacedMissingFontLine(h)
			h = ffmpeg.KillOnGlyphNotFoundLine(h)
		}
		h = ffmpeg.KillOnNotOverwritingLine(h)
		for s.Scan() {
			line := s.Text()
			h.Handle(commandline.Response{Signaller: cmd.Process, Stdout: out}, line)
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

// modifiableOutput is an io.Writer which can detect carriage return
// and there for allow for a line to be modified.
type modifiableOutput struct {
	Stdout            io.Writer
	hasCarriageReturn bool
}

func (f *modifiableOutput) Write(p []byte) (n int, err error) {
	if i := bytes.LastIndex(p, []byte("\r")); i > -1 && i > bytes.LastIndex(p, []byte("\n")) {
		f.hasCarriageReturn = true
		return f.Stdout.Write(p)
	}
	if f.hasCarriageReturn {
		f.hasCarriageReturn = false
		p = append([]byte("\n"), p...)
	}
	if !bytes.HasSuffix(p, []byte("\n")) {
		p = append(p, []byte("\n")...)
	}
	return f.Stdout.Write(p)
}
