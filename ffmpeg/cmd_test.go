package ffmpeg

import (
	"github.com/shiroi-usagi/burner/commandline"
	"io/ioutil"
	"os"
	"testing"
)

type fakeSignaller struct {
	lastSignal os.Signal
}

func (f *fakeSignaller) Signal(signal os.Signal) error {
	f.lastSignal = signal
	return nil
}

func (f *fakeSignaller) reset() {
	f.lastSignal = nil
}

func TestKillOnGlyphNotFoundLine(t *testing.T) {
	signaller := fakeSignaller{}
	r := commandline.Response{Signaller: &signaller, Stdout: ioutil.Discard}
	var calledNext bool
	pipe := KillOnGlyphNotFoundLine(commandline.HandlerFunc(func(_ commandline.Response, _ string) {
		calledNext = true
	}))

	calledNext = false
	signaller.reset()
	pipe.Handle(r, `[Parsed_subtitles_0 @ anyhex] Glyph 0x266F not found, selecting one more font for (anystring, 0, 0)`)
	if signaller.lastSignal != os.Kill {
		t.Errorf("should call kill signal on Glyph not found warning")
	}
	if calledNext != false {
		t.Errorf("should not call next on Glyph not found warning")
	}

	calledNext = false
	signaller.reset()
	pipe.Handle(r, `any line`)
	if signaller.lastSignal == os.Kill {
		t.Errorf("should not call kill signal on safe string")
	}
	if calledNext != true {
		t.Errorf("should call next on safe string")
	}
}

func TestKillOnNotOverwritingLine(t *testing.T) {
	signaller := fakeSignaller{}
	r := commandline.Response{Signaller: &signaller, Stdout: ioutil.Discard}
	var calledNext bool
	pipe := KillOnNotOverwritingLine(commandline.HandlerFunc(func(_ commandline.Response, _ string) {
		calledNext = true
	}))

	calledNext = false
	signaller.reset()
	pipe.Handle(r, `File 'anyfile' already exists. Overwrite ? [y/N] Not overwriting - exiting`)
	if signaller.lastSignal != os.Kill {
		t.Errorf("should call kill signal on Glyph not found warning")
	}
	if calledNext != false {
		t.Errorf("should not call next on Glyph not found warning")
	}

	calledNext = false
	signaller.reset()
	pipe.Handle(r, `any line`)
	if signaller.lastSignal == os.Kill {
		t.Errorf("should not call kill signal on safe string")
	}
	if calledNext != true {
		t.Errorf("should call next on safe string")
	}
}

func TestKillOnReplacedMissingFontLine(t *testing.T) {
	tests := []struct {
		name        string
		commandline string
		wantSignal  os.Signal
		wantNext    bool
	}{
		{
			name:        "font replaced with Arial",
			commandline: "[Parsed_subtitles_0 @ anyhex] fontselect: (HZsH_Xirwena, 400, 0) -> ArialMT, 0, ArialMT",
			wantSignal:  os.Kill,
		},
		{
			name:        "higher id font replaced with Arial",
			commandline: "[Parsed_subtitles_1 @ 000002475be4c140] fontselect: (Teszt1, 400, 0) -> ArialMT, 0, ArialMT",
			wantSignal:  os.Kill,
		},
		{
			name:        "third option is minus with Arial",
			commandline: "[Parsed_subtitles_1 @ 000002475be4c140] fontselect: (Teszt1, 400, 0) -> ArialMT, -1, ArialMT",
			wantSignal:  os.Kill,
		},
		{
			name:        "font replaced with Arial Bold",
			commandline: "[Parsed_subtitles_0 @ 000002475be4c140] fontselect: (Teszt2, 700, 0) -> Arial-BoldMT, 0, Arial-BoldMT",
			wantSignal:  os.Kill,
		},
		{
			name:        "font replaced with DejaVuSans",
			commandline: "[Parsed_subtitles_0 @ anyhex] fontselect: (HZsH_Xirwena, 400, 0) -> /font/path/DejaVuSans.ttf, 0, DejaVuSans",
			wantSignal:  os.Kill,
		},
		{
			name:        "font replaced with Helvetica",
			commandline: "[Parsed_subtitles_0 @ anyhex] fontselect: (HZsH_Xirwena, 400, 0) -> /font/path/DejaVuSans.ttf, 0, Helvetica",
			wantSignal:  os.Kill,
		},
		{
			name:        "font is Arial",
			commandline: "[Parsed_subtitles_0 @ anyhex] fontselect: (Arial, 400, 0) -> ArialMT, 0, ArialMT",
			wantNext:    true,
		},
		{
			name:        "font is DejaVuSans",
			commandline: "[Parsed_subtitles_0 @ anyhex] fontselect: (DejaVuSans, 400, 0) -> /font/path/DejaVuSans.ttf, 0, DejaVuSans",
			wantNext:    true,
		},
		{
			name:        "font is Helvetica",
			commandline: "[Parsed_subtitles_0 @ anyhex] fontselect: (Helvetica, 400, 0) -> /font/path/Helvetica.ttf, 0, Helvetica",
			wantNext:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signaller := fakeSignaller{}
			r := commandline.Response{Signaller: &signaller, Stdout: ioutil.Discard}
			var calledNext bool
			pipe := KillOnReplacedMissingFontLine(commandline.HandlerFunc(func(_ commandline.Response, _ string) {
				calledNext = true
			}))
			pipe.Handle(r, tt.commandline)
			if got := signaller.lastSignal; got != tt.wantSignal {
				t.Errorf("KillOnReplacedMissingFontLine() = %v, wantSignal %v", got, tt.wantSignal)
			}
			if got := calledNext; got != tt.wantNext {
				t.Errorf("KillOnReplacedMissingFontLine() = %v, wantSignal %v", got, tt.wantNext)
			}
		})
	}
}
