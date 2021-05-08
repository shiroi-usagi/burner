package ffmpeg

import (
	"fmt"
	"github.com/shiroi-usagi/burner/commandline"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const commandlineStatusPrefix = "frame="

func StatusPrinter() (h commandline.Handler) {
	return commandline.HandlerFunc(func(r commandline.Response, l string) {
		if strings.HasPrefix(l, commandlineStatusPrefix) {
			fmt.Fprint(r.Stdout, l+"\r")
		}
	})
}

func Printer() commandline.Handler {
	return commandline.HandlerFunc(func(r commandline.Response, l string) {
		fmt.Fprint(r.Stdout, l)
	})
}

// The pattern for glyphNotFoundMatcher detects all three possible Glyph errors.
//
// https://github.com/libass/libass/blob/81e99a73d16873a782c99d068db99485043fcba4/libass/ass_font.c#L473-L499
var glyphNotFoundMatcher = regexp.MustCompile(`^\[Parsed_subtitles_\d+ @ \w+] Glyph 0x(\w+) not found`)

// KillOnGlyphNotFoundLine detects Glyph not found errors
// in the ffmpeg output. When the error happens stops propagation.
func KillOnGlyphNotFoundLine(next commandline.Handler) commandline.Handler {
	return commandline.HandlerFunc(func(r commandline.Response, l string) {
		match := glyphNotFoundMatcher.FindStringSubmatch(l)
		if len(match) > 0 {
			i, err := strconv.ParseInt(match[1], 16, 64)
			if err != nil {
				// Not expecting to have invalid hex
				panic(err)
			}
			_ = r.Signal(os.Kill)
			fmt.Fprintf(r.Stdout, "burner: was not able to find font for `%s` char", string(rune(i)))
			return
		}
		next.Handle(r, l)
	})
}

// KillOnNotOverwritingLine overwriting errors in the ffmpeg output.
// When the error happens stops propagation.
func KillOnNotOverwritingLine(next commandline.Handler) commandline.Handler {
	return commandline.HandlerFunc(func(r commandline.Response, l string) {
		if strings.HasSuffix(l, "Not overwriting - exiting") {
			i := strings.LastIndex(l, ".") + 1
			_ = r.Signal(os.Kill)
			fmt.Fprintf(r.Stdout, "burner: %s", l[:i])
			return
		}
		next.Handle(r, l)
	})
}

// fontReplacementMatchers contains matchers for detecting
// font replacements from ffmpeg commandline output.
var fontReplacementMatchers = map[string]*regexp.Regexp{
	// libass replaces a missing font with Arial when a default font is not specified
	"Arial": regexp.MustCompile(`\[Parsed_subtitles_\d+ @ \w+] fontselect: \((.*?), \d+, \d+\) -> .*?, -?\d+, (?:ArialMT|Arial-BoldMT|Arial-ItalicMT)$`),

	// libass replaces a missing font with DejaVuSans on Linux when a default font is not specified
	"DejaVuSans": regexp.MustCompile(`\[Parsed_subtitles_\d+ @ \w+] fontselect: \((.*?), \d+, \d+\) -> .*?, -?\d+, (?:DejaVuSans)$`),

	// libass replaces a missing font with Helvetica on MacOS when a default font is not specified
	"Helvetica": regexp.MustCompile(`\[Parsed_subtitles_\d+ @ \w+] fontselect: \((.*?), \d+, \d+\) -> .*?, -?\d+, (?:Helvetica)$`),
}

// KillOnNotOverwritingLine overwriting errors in the ffmpeg output.
// When the error happens stops propagation.
func KillOnReplacedMissingFontLine(next commandline.Handler) commandline.Handler {
	return commandline.HandlerFunc(func(r commandline.Response, l string) {
		for prefix, matcher := range fontReplacementMatchers {
			match := matcher.FindStringSubmatch(l)
			if len(match) > 0 && !strings.HasPrefix(match[1], prefix) {
				_ = r.Signal(os.Kill)
				fmt.Fprintf(r.Stdout, "burner: missing `%s` font", match[1])
				return
			}
		}
		next.Handle(r, l)
	})
}
