package burner

import (
	"bufio"
	"bytes"
	"container/ring"
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
	// For some reason FFmpeg writes to stderr
	e, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if conf.Verbose {
		fmt.Println(cmd)
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(e)
	scanner.Split(ScanLineAndCarriageReturn)
	cbuf := ring.New(5)
	h := ffmpeg.StatusPrinter()
	if conf.Verbose {
		h = ffmpeg.Printer()
	}
	if !conf.IgnoreFontError {
		h = ffmpeg.KillOnReplacedMissingFontLine(h)
		h = ffmpeg.KillOnGlyphNotFoundLine(h)
	}
	h = ffmpeg.KillOnNotOverwritingLine(h)
	for scanner.Scan() {
		m := scanner.Text()
		cbuf.Value = m
		cbuf = cbuf.Next()
		h.Handle(commandline.Response{Signaller: cmd.Process, Stdout: out}, m)
	}

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.Exited() {
		cbuf.Do(func(i interface{}) {
			if i != nil {
				fmt.Println(fmt.Sprintf("%s", i))
			}
		})
	}
	return err
}

// ScanLineAndCarriageReturn is a split function for a Scanner that returns
// each line of text, stripped of any trailing end-of-line marker or carriage
// return. The returned line may be empty. The end-of-line marker is one
// optional carriage return followed by one mandatory newline. In regular
// expression notation, it is `(\r|\r?\n)`. The last non-empty line of input
// will be returned even if it has no newline.
//
// Carriage return is used when a command overrides a single line of output.
func ScanLineAndCarriageReturn(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
//
// Borrowed from bufio/scan
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
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
