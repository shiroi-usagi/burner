package ffmpeg

import (
	"testing"
)

func TestFilter_String(t *testing.T) {
	type fields struct {
		subtitle  string
		width     int
		height    int
		upscaling bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "subtitle",
			fields: fields{subtitle: `/in/file.mkv`},
			want:   `subtitles='/in/file.mkv'`,
		},
		{
			name:   "subtitle escaped",
			fields: fields{subtitle: `C:\in\file.mkv`},
			want:   `subtitles='C\:\\in\\file.mkv'`,
		},
		{
			name:   "scale",
			fields: fields{width: -1, height: 720, upscaling: true},
			want:   `scale=-1:720`,
		},
		{
			name:   "scale with avoiding upscaling",
			fields: fields{width: 320, height: 240},
			want:   `scale='min(320,iw)':'min(240,ih)'`,
		},
		{
			name:   "concatenate filters",
			fields: fields{subtitle: `/in/file.mkv`, width: 320, height: 240, upscaling: true},
			want:   `subtitles='/in/file.mkv', scale=320:240`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filter{
				Subtitle:  tt.fields.subtitle,
				Width:     tt.fields.width,
				Height:    tt.fields.height,
				Upscaling: tt.fields.upscaling,
			}
			if got := f.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitrateToKilobit(t *testing.T) {
	type args struct {
		bitrate string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "kilobit",
			args: args{bitrate: "1371k"},
			want: 1371,
		},
		{
			name: "megabit",
			args: args{bitrate: "1M"},
			want: 1024,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BitrateToKilobit(tt.args.bitrate); got != tt.want {
				t.Errorf("BitrateToKilobit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKilobitToBitrate(t *testing.T) {
	type args struct {
		kilobit int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "kilobit",
			args: args{kilobit: 1371},
			want: "1371k",
		},
		{
			name: "megabit",
			args: args{kilobit: 1024},
			want: "1024k",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KilobitToBitrate(tt.args.kilobit); got != tt.want {
				t.Errorf("BitrateToKilobit() = %v, want %v", got, tt.want)
			}
		})
	}
}
