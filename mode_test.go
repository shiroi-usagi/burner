package burner

import (
	"io"
	"strings"
	"testing"
)

func TestStringToMode(t *testing.T) {
	type args struct {
		m string
	}
	tests := []struct {
		name string
		args args
		want Mode
	}{
		{
			name: "invalid mode",
			args: args{m: ""},
			want: ModeNone,
		},
		{
			name: "existing mode",
			args: args{m: "smp4"},
			want: ModeSampleMP4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToMode(tt.args.m); got != tt.want {
				t.Errorf("StringToMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadMode(t *testing.T) {
	type args struct {
		reader io.RuneReader
	}
	tests := []struct {
		name string
		args args
		want Mode
	}{
		{
			name: "empty string",
			args: args{reader: strings.NewReader("")},
			want: ModeNone,
		},
		{
			name: "invalid num",
			args: args{reader: strings.NewReader("a")},
			want: ModeNone,
		},
		{
			name: "invalid mode",
			args: args{reader: strings.NewReader("9")},
			want: ModeNone,
		},
		{
			name: "existing mode",
			args: args{reader: strings.NewReader("0")},
			want: ModeSampleMP4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadMode(tt.args.reader); got != tt.want {
				t.Errorf("ReadMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
